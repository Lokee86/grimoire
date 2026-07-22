# frozen_string_literal: true

class LexiconRubyAdapter
  VERSION = LexiconRuby::Contract::VERSION
  LANGUAGE = LexiconRuby::Contract::LANGUAGE
  EXCLUDED_DIRECTORIES = LexiconRuby::Contract::EXCLUDED_DIRECTORIES
  METAPROGRAMMING_CALLS = LexiconRuby::Contract::METAPROGRAMMING_CALLS

  include LexiconRuby::Contract
  include LexiconRuby::Relationships
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
