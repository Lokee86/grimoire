# frozen_string_literal: true

class LexiconRubyAdapter
  VERSION = LexiconRuby::Contract::VERSION
  LANGUAGE = LexiconRuby::Contract::LANGUAGE
  EXCLUDED_DIRECTORIES = LexiconRuby::Contract::EXCLUDED_DIRECTORIES
  METAPROGRAMMING_CALLS = LexiconRuby::Contract::METAPROGRAMMING_CALLS

  include LexiconRuby::Contract
  include LexiconRuby::Relationships
  include LexiconRuby::RipperSyntax
  include LexiconRuby::RipperExtractor
  include LexiconRuby::RepositoryDiscovery
  include LexiconRuby::JsonlEmitter

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
    @method_definitions = Hash.new { |hash, key| hash[key] = Hash.new { |inner, name| inner[name] = [] } }
    @method_contexts = {}
    @pending_calls = []
    @processed_calls = {}
    @source_lines = []
    @current_path = nil
  end

  private

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
