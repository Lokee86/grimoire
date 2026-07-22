#!/usr/bin/env ruby
# frozen_string_literal: true

require "digest"
require "fileutils"
require "json"
require "optparse"
require "ripper"

# First runnable Ruby Lexicon adapter slice.
#
# The parser boundary is deliberately Ripper-based. The rest of the adapter is
# independent of that choice so a richer parser can replace parse_file later.
class LexiconRubyAdapter
  VERSION = "0.1.0"
  LANGUAGE = "ruby"
  EXCLUDED_DIRECTORIES = %w[
    .git .worktrees .workingtrees .warlock
    .bundle vendor node_modules target build dist tmp log coverage
  ].freeze
  METAPROGRAMMING_CALLS = %w[
    class_eval module_eval instance_eval define_method define_singleton_method
    const_set class_variable_set eval instance_exec send public_send
  ].freeze

  def initialize(repo, output)
    @repo = File.expand_path(repo)
    @output = File.expand_path(output)
    @repository_name = File.basename(@repo)
    @nodes = {}
    @directories = {}
    @files = {}
    @edges = {}
    @unresolved = []
    @type_nodes = Hash.new { |hash, key| hash[key] = [] }
    @pending_extends = []
    @processed_calls = {}
    @source_lines = []
    @current_path = nil
  end

  def run
    raise ArgumentError, "repository does not exist: #{@repo}" unless Dir.exist?(@repo)

    repository_id = add_node(
      kind: "repository",
      name: @repository_name,
      path: ".",
      qualified_name: @repository_name,
      canonical: @repository_name
    )
    root_directory_id = add_directory(".")
    add_edge(repository_id, root_directory_id, "contains")

    ruby_files.each do |relative_path|
      scan_file(relative_path)
    end

    resolve_superclasses
    write_facts
  end

  private

  def ruby_files
    Dir.glob(File.join(@repo, "**", "*.rb"), File::FNM_EXTGLOB).filter_map do |absolute_path|
      next unless File.file?(absolute_path)

      relative_path = normalize_path(Pathname_relative(absolute_path))
      next if excluded_path?(relative_path)

      relative_path
    end.sort
  end

  # Avoid requiring Pathname just to keep this adapter's dependency surface
  # explicit and standard-library-only.
  def Pathname_relative(path)
    path.delete_prefix("#{@repo}#{File::SEPARATOR}").delete_prefix("#{@repo}/")
  end

  def normalize_path(path)
    path.tr("\\", "/")
  end

  def excluded_path?(relative_path)
    relative_path.split("/").any? { |component| EXCLUDED_DIRECTORIES.include?(component) }
  end

  def scan_file(relative_path)
    absolute_path = File.join(@repo, relative_path)
    source = File.binread(absolute_path).force_encoding(Encoding::UTF_8)
    @source_lines = source.lines
    @current_path = relative_path

    file_id = add_node(
      kind: "file",
      name: File.basename(relative_path),
      path: relative_path,
      qualified_name: relative_path,
      canonical: relative_path,
      content_id: content_id(source)
    )
    @files[relative_path] = file_id
    parent_directory_id = add_directory(File.dirname(relative_path))
    add_edge(parent_directory_id, file_id, "contains")

    sexp = Ripper.sexp(source)
    if sexp.nil?
      add_unresolved(
        source: file_id,
        relation: "parses",
        expression: relative_path,
        reason: "unsupported-form"
      )
      return
    end

    visit(sexp, file_id, "")
  rescue ArgumentError => error
    add_unresolved(
      source: @files[relative_path] || add_node(
        kind: "file",
        name: File.basename(relative_path),
        path: relative_path,
        qualified_name: relative_path,
        canonical: relative_path
      ),
      relation: "parses",
      expression: relative_path,
      reason: "unsupported-form",
      attributes: { "message" => error.message }
    )
  end

  def add_directory(relative_path)
    relative_path = normalize_path(relative_path)
    return @directories[relative_path] if @directories.key?(relative_path)

    directory_id = add_node(
      kind: "directory",
      name: relative_path == "." ? @repository_name : File.basename(relative_path),
      path: relative_path,
      qualified_name: relative_path,
      canonical: relative_path
    )
    @directories[relative_path] = directory_id

    unless relative_path == "."
      parent = File.dirname(relative_path)
      parent = "." if parent == ""
      parent_id = add_directory(parent)
      add_edge(parent_id, directory_id, "contains")
    end

    directory_id
  end

  def visit(node, parent_id, namespace)
    return unless node.is_a?(Array)

    case node[0]
    when :program
      visit_body(node[1], parent_id, namespace)
    when :module
      visit_module(node, parent_id, namespace)
    when :class
      visit_class(node, parent_id, namespace)
    when :def
      visit_method(node, parent_id, namespace, singleton: false)
    when :defs
      visit_method(node, parent_id, namespace, singleton: true)
    when :assign
      visit_assignment(node, parent_id, namespace)
    when :command
      process_call(node[1], node[2], parent_id, namespace)
      visit(node[2], parent_id, namespace)
    when :command_call
      process_call(node[3], node[4], parent_id, namespace)
      visit(node[1], parent_id, namespace)
      visit(node[4], parent_id, namespace)
    when :method_add_arg
      call = node[1]
      process_call(call_name_token(call), node[2], parent_id, namespace)
      visit(call, parent_id, namespace)
      visit(node[2], parent_id, namespace)
    when :sclass
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: expression_for(node),
        reason: "unsupported-form",
        span: span_for(first_token(node))
      )
      visit_body(node[2], parent_id, namespace)
    when :alias, :undef
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: expression_for(node),
        reason: "unsupported-form",
        span: span_for(first_token(node))
      )
    else
      node[1..].each { |child| visit(child, parent_id, namespace) }
    end
  end

  def visit_body(body, parent_id, namespace)
    return unless body

    if sexp_node?(body, :bodystmt)
      body[1..].each { |part| visit(part, parent_id, namespace) }
    elsif body.is_a?(Array) && body.first.is_a?(Array)
      body.each { |part| visit(part, parent_id, namespace) }
    else
      visit(body, parent_id, namespace)
    end
  end

  def visit_module(node, parent_id, namespace)
    name = const_name(node[1])
    unless name
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: expression_for(node[1]),
        reason: "unsupported-form",
        span: span_for(first_token(node))
      )
      return
    end

    qualified_name = qualify(namespace, name)
    module_id = add_symbol("module", name.split("::").last, qualified_name, node[1], {})
    connect_declaration(parent_id, module_id, span_for(first_token(node[1])))
    visit_body(node[2], module_id, qualified_name)
  end

  def visit_class(node, parent_id, namespace)
    name = const_name(node[1])
    unless name
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: expression_for(node[1]),
        reason: "unsupported-form",
        span: span_for(first_token(node))
      )
      return
    end

    qualified_name = qualify(namespace, name)
    superclass = const_name(node[2])
    attributes = {}
    if superclass
      attributes["superclass"] = qualify(namespace, superclass)
    end
    class_id = add_symbol("type", name.split("::").last, qualified_name, node[1], attributes)
    @type_nodes[qualified_name] << class_id
    connect_declaration(parent_id, class_id, span_for(first_token(node[1])))

    if node[2]
      if superclass
        @pending_extends << [class_id, qualify(namespace, superclass), span_for(first_token(node[2]))]
      else
        add_unresolved(
          source: class_id,
          relation: "extends",
          expression: expression_for(node[2]),
          reason: "unsupported-form",
          span: span_for(first_token(node[2]))
        )
      end
    end

    visit_body(node[3], class_id, qualified_name)
  end

  def visit_method(node, parent_id, namespace, singleton:)
    name_token = singleton ? node[2] : node[1]
    method_name = token_value(name_token)
    unless method_name
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: expression_for(node),
        reason: "unsupported-form",
        span: span_for(first_token(node))
      )
      return
    end

    owner_name = namespace.empty? ? "" : namespace
    qualified_name = if singleton
                      receiver = const_name(node[1])
                      receiver = qualify(namespace, receiver) if receiver
                      receiver ? "#{receiver}.#{method_name}" : "#{owner_name}.#{method_name}".delete_prefix(".")
                    else
                      owner_name.empty? ? method_name : "#{owner_name}##{method_name}"
                    end
    attributes = { "singleton" => singleton }
    parameters = parameter_names(singleton ? node[3] : node[2])
    attributes["parameters"] = parameters unless parameters.empty?
    method_id = add_symbol("method", method_name, qualified_name, name_token, attributes)
    connect_declaration(parent_id, method_id, span_for(name_token))
    visit_body(singleton ? node[4] : node[3], method_id, namespace)
  end

  def visit_assignment(node, parent_id, namespace)
    name = const_name(node[1])
    if name
      qualified_name = qualify(namespace, name)
      constant_id = add_symbol("constant", name.split("::").last, qualified_name, node[1], {})
      connect_declaration(parent_id, constant_id, span_for(first_token(node[1])))
    end
    visit(node[2], parent_id, namespace)
  end

  def process_call(name_token, arguments, parent_id, namespace)
    name = token_value(name_token)
    return unless name

    call_key = [@current_path, name, token_position(name_token)]
    return if @processed_calls[call_key]

    @processed_calls[call_key] = true
    if %w[require require_relative load].include?(name)
      required = literal_string(first_argument(arguments))
      if required
        import_id = add_symbol(
          "import",
          required,
          "#{@current_path}:#{required}:#{token_position(name_token).join(":")}",
          name_token,
          { "loader" => name, "target" => required },
          canonical_extra: "#{@current_path}\0#{token_position(name_token).join(":")}\0#{required}"
        )
        connect_declaration(parent_id, import_id, span_for(name_token))
        add_edge(parent_id, import_id, "imports", span_for(name_token))
      else
        add_unresolved(
          source: parent_id,
          relation: "imports",
          expression: expression_for(arguments),
          reason: "dynamic-target",
          span: span_for(name_token)
        )
      end
    elsif METAPROGRAMMING_CALLS.include?(name)
      add_unresolved(
        source: parent_id,
        relation: "defines",
        expression: name,
        reason: "dynamic-target",
        span: span_for(name_token),
        attributes: { "namespace" => namespace }
      )
    end
  end

  def resolve_superclasses
    @pending_extends.each do |source_id, qualified_name, span|
      target_id = @type_nodes[qualified_name].first
      target_id ||= add_node(
        kind: "type",
        name: qualified_name.split("::").last,
        path: "<external>",
        qualified_name: qualified_name,
        canonical: "external\0#{qualified_name}",
        attributes: { "external" => true }
      )
      add_edge(source_id, target_id, "extends", span)
    end
  end

  def connect_declaration(parent_id, child_id, span)
    add_edge(parent_id, child_id, "contains", span)
    add_edge(parent_id, child_id, "defines", span)
  end

  def add_symbol(kind, name, qualified_name, token, attributes, canonical_extra: nil)
    path = @current_path || "."
    canonical = [path, qualified_name, canonical_extra].compact.join("\0")
    add_node(
      kind: kind,
      name: name,
      path: path,
      qualified_name: qualified_name,
      canonical: canonical,
      attributes: attributes,
      span: span_for(token)
    )
  end

  def add_node(kind:, name:, path:, qualified_name:, canonical:, attributes: nil, span: nil, content_id: nil)
    id = node_id(kind, canonical)
    record = {
      "record" => "node",
      "id" => id,
      "kind" => kind,
      "name" => name,
      "path" => path,
      "qualified_name" => qualified_name
    }
    record["content_id"] = content_id if content_id
    record["span"] = span if span
    record["attributes"] = attributes unless attributes.nil? || attributes.empty?
    @nodes[id] ||= record
    id
  end

  def add_edge(source, target, relation, span = nil)
    record = { "record" => "edge", "source" => source, "target" => target, "relation" => relation }
    record["span"] = span if span
    @edges[canonical_record(record)] = record
  end

  def add_unresolved(source:, relation:, expression:, reason:, span: nil, attributes: nil)
    record = {
      "record" => "unresolved",
      "source" => source,
      "relation" => relation,
      "expression" => expression.to_s,
      "reason" => reason
    }
    record["span"] = span if span
    record["attributes"] = attributes if attributes && !attributes.empty?
    @unresolved << record
  end

  def write_facts
    header = {
      "record" => "lexicon",
      "schema_version" => 1,
      "adapter_version" => VERSION,
      "language" => LANGUAGE,
      "repository" => @repository_name
    }
    records = @nodes.values.sort_by { |record| [record["id"], record["kind"], record["path"], record["qualified_name"]] }
    records += @edges.values.sort_by { |record| [record["source"], record["target"], record["relation"], *span_key(record)] }
    records += @unresolved.sort_by do |record|
      [record["source"], record["relation"], record["expression"], record["reason"], *span_key(record)]
    end

    output_directory = File.dirname(@output)
    FileUtils.mkdir_p(output_directory) unless output_directory == "."
    File.open(@output, "wb") do |file|
      ([header] + records).each { |record| file.write("#{canonical_json(record)}\n") }
    end
    @output
  end

  def node_id(kind, canonical)
    "sha256:#{Digest::SHA256.hexdigest("lexicon:v1\0#{LANGUAGE}\0#{kind}\0#{canonical}")}"
  end

  def content_id(content)
    "sha256:#{Digest::SHA256.hexdigest(content)}"
  end

  def sexp_node?(value, tag)
    value.is_a?(Array) && value.first == tag
  end

  def token?(value)
    value.is_a?(Array) && value.length == 3 && value.first.is_a?(Symbol) && value.first.to_s.start_with?("@")
  end

  def first_token(value)
    return value if token?(value)
    return nil unless value.is_a?(Array)

    value.each do |child|
      token = first_token(child)
      return token if token
    end
    nil
  end

  def token_value(token)
    token?(token) ? token[1] : nil
  end

  def token_position(token)
    token?(token) ? token[2] : [0, 0]
  end

  def const_name(value)
    return nil unless value
    return value[1] if token?(value) && value[0] == :@const

    if value.is_a?(Array)
      case value[0]
      when :const_ref, :var_ref, :var_field, :top_const_ref
        name = const_name(value[1])
        return name && (value[0] == :top_const_ref ? "::#{name}" : name)
      when :const_path_ref, :const_path_field
        left = const_name(value[1])
        right = const_name(value[2])
        return "#{left}::#{right}" if left && right
      end
    end
    nil
  end

  def qualify(namespace, name)
    name = name.to_s.delete_prefix("::")
    return name if namespace.empty? || name.include?("::")

    "#{namespace}::#{name}"
  end

  def parameter_names(params)
    tokens = []
    collect_tokens(params).each do |token|
      tokens << token[1] if %i[@ident @const].include?(token[0])
    end
    tokens.uniq
  end

  def collect_tokens(value, result = [])
    if token?(value)
      result << value
    elsif value.is_a?(Array)
      value.each { |child| collect_tokens(child, result) }
    end
    result
  end

  def first_argument(arguments)
    return nil unless arguments.is_a?(Array)
    return arguments[1].is_a?(Array) ? arguments[1].first : arguments[1] if arguments[0] == :args_add_block
    return first_argument(arguments[1]) if arguments[0] == :arg_paren

    arguments
  end

  def literal_string(value)
    return nil unless value.is_a?(Array)
    return value[1] if value[0] == :@tstring_content

    if value[0] == :string_literal
      content = value[1]
      return nil unless content.is_a?(Array) && content[0] == :string_content

      parts = content[1..].to_a
      return nil if parts.any? { |part| part.is_a?(Array) && part[0] == :string_embexpr }

      text = parts.filter_map { |part| part[1] if part.is_a?(Array) && part[0] == :@tstring_content }.join
      return text unless text.empty? && parts.any?
    end
    nil
  end

  def call_name_token(call)
    return nil unless call.is_a?(Array)

    case call[0]
    when :fcall, :vcall
      call[1]
    when :call
      call[3]
    when :command_call
      call[3]
    when :command
      call[1]
    end
  end

  def expression_for(value)
    tokens = collect_tokens(value)
    return "" if tokens.empty?

    tokens.map { |token| token[1].to_s }.join
  end

  def span_for(token)
    return nil unless token?(token)

    line, column = token[2]
    text = token[1].to_s
    {
      "start_line" => line,
      "start_column" => column + 1,
      "end_line" => line,
      "end_column" => column + text.length + 1,
      "path" => @current_path
    }
  end

  def span_key(record)
    span = record["span"] || {}
    [span["path"] || "", span["start_line"] || 0, span["start_column"] || 0,
     span["end_line"] || 0, span["end_column"] || 0]
  end

  def canonical_record(record)
    canonical_json(record)
  end

  def canonical_json(value)
    JSON.generate(sort_hashes(value))
  end

  def sort_hashes(value)
    case value
    when Hash
      value.keys.sort.each_with_object({}) { |key, result| result[key] = sort_hashes(value[key]) }
    when Array
      value.map { |item| sort_hashes(item) }
    else
      value
    end
  end
end

if __FILE__ == $PROGRAM_NAME
  options = {}
  parser = OptionParser.new do |opts|
    opts.banner = "Usage: lexicon_ruby.rb --repo PATH --output PATH"
    opts.on("--repo PATH", "repository root") { |value| options[:repo] = value }
    opts.on("--output PATH", "JSONL output path") { |value| options[:output] = value }
  end

  begin
    parser.parse!
    missing = %i[repo output].reject { |key| options[key] }
    abort("missing required option(s): #{missing.map { |key| "--#{key}" }.join(", ")}\n#{parser}") unless missing.empty?
    LexiconRubyAdapter.new(options[:repo], options[:output]).run
  rescue OptionParser::ParseError, ArgumentError => error
    abort(error.message)
  end
end
