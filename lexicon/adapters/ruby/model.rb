# frozen_string_literal: true

class LexiconRubyAdapter
  VERSION = LexiconRuby::Contract::VERSION
  LANGUAGE = LexiconRuby::Contract::LANGUAGE
  EXCLUDED_DIRECTORIES = LexiconRuby::Contract::EXCLUDED_DIRECTORIES
  METAPROGRAMMING_CALLS = LexiconRuby::Contract::METAPROGRAMMING_CALLS

  include LexiconRuby::Contract
  include LexiconRuby::Relationships
  include LexiconRuby::DependencySemantics
  include LexiconRuby::RipperSyntax
  include LexiconRuby::RipperClosures
  include LexiconRuby::RipperCalls
  include LexiconRuby::RipperDeclarations
  include LexiconRuby::RipperAssignments
  include LexiconRuby::RipperExtractor
  include LexiconRuby::CallShapes
  include LexiconRuby::CallHierarchy
  include LexiconRuby::CallFlow
  include LexiconRuby::CallEmission
  include LexiconRuby::CallResolution
  include LexiconRuby::RepositoryDiscovery
  include LexiconRuby::JsonlEmitter

  def initialize(repo, output, changed_files = nil, removed_files = nil)
    @repo = File.expand_path(repo)
    @output = File.expand_path(output)
    @changed_files = changed_files&.map { |path| normalize_emission_path(path) }
    @removed_files = removed_files&.map { |path| normalize_emission_path(path) }
    @repository_name = File.basename(@repo)
    @nodes = {}
    @directories = {}
    @files = {}
    @module_ids = {}
    @pending_local_dependencies = []
    @edges = {}
    @unresolved = []
    @types = {}
    @pending_extends = []
    @pending_mixin_edges = []
    @methods = {}
    @method_contexts = {}
    @method_definitions = Hash.new do |owners, owner|
      owners[owner] = Hash.new do |modes, mode|
        modes[mode] = Hash.new { |names, name| names[name] = [] }
      end
    end
    @constructors = {}
    @aliases = {}
    @method_alias_targets = {}
    @module_function_all = Set.new
    @undefined_methods = Hash.new { |hash, key| hash[key] = Set.new }
    @mixin_hosts = Hash.new { |hash, key| hash[key] = Set.new }
    @call_sites = {}
    @call_node_keys = {}
    @resolved_targets = {}
    @method_returns = {}
    @parameter_shapes = {}
    @assignments = []
    @data_symbols = {}
    @dataflow_edges = Set.new
    @constant_assignments = {}
    @constant_type_assignments = {}
    @lambda_ids = {}
    @blocks = {}
    @scope_parents = {}
    @passed_blocks = Hash.new { |hash, key| hash[key] = Set.new }
    @source_lines = []
    @current_path = nil
  end

  private

  def normalize_emission_path(path)
    path.tr("\\\\", "/")
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

  def data_symbol_id(kind, name, context, token = nil)
    normalized_kind = kind.to_sym
    node_kind = normalized_kind == :constant ? "constant" : (normalized_kind == :parameter ? "parameter" : "variable")
    scope = %i[instance class_variable].include?(normalized_kind) ? context[:owner].to_s : context[:scope_id].to_s
    return nil if scope.empty? || name.to_s.empty?

    key = [scope, normalized_kind, name.to_s]
    return @data_symbols[key] if @data_symbols[key]

    owner_name = @nodes[scope]&.fetch("qualified_name", context[:owner].to_s) || @current_path
    qualified_name = "#{owner_name}.#{name}"
    id = add_node(
      kind: node_kind,
      name: name.to_s,
      path: @current_path,
      qualified_name: qualified_name,
      canonical: "data\0#{scope}\0#{normalized_kind}\0#{name}",
      span: span_for(token)
    )
    @data_symbols[key] = id
    add_edge(scope, id, "defines", span_for(token)) if @nodes[scope]
    id
  end

  def resolve_data_symbol(kind, name, context)
    normalized_kind = kind.to_sym
    scopes = if %i[instance class_variable].include?(normalized_kind)
               [context[:owner].to_s]
             else
               result = []
               current = context[:scope_id].to_s
               while !current.empty?
                 result << current
                 current = @scope_parents[current].to_s
               end
               result
             end
    found = scopes.filter_map do |scope|
      @data_symbols[[scope, normalized_kind, name.to_s]] ||
        (normalized_kind == :local ? @data_symbols[[scope, :parameter, name.to_s]] : nil)
    end.first
    return found if found || normalized_kind != :constant

    @data_symbols.each do |(scope, kind, symbol_name), symbol_id|
      next unless kind == :constant && symbol_name == name.to_s
      qualified = @nodes[symbol_id]&.fetch("qualified_name", "").to_s
      owner_qname = @nodes[context[:owner].to_s]&.fetch("qualified_name", "").to_s
      return symbol_id if owner_qname.empty? || qualified.start_with?("#{owner_qname}::")
    end
    nil
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
end
