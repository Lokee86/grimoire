# frozen_string_literal: true

module LexiconRuby
  module Relationships
    private

    def add_edge(source, target, relation, span = nil, attributes = nil)
      return unless source && target

      record = { "record" => "edge", "source" => source, "target" => target, "relation" => relation }
      record["span"] = span if span
      record["attributes"] = attributes if attributes && !attributes.empty?
      span_key = span&.values_at("path", "start_line", "start_column", "end_line", "end_column")
      attribute_key = attributes&.sort
      @edges[[source, target, relation, span_key, attribute_key]] ||= record
    end

    def add_dataflow_edge(source, target, relation, span = nil)
      key = [source, target, relation]
      return if @dataflow_edges.include?(key)

      @dataflow_edges << key
      add_edge(source, target, relation, span)
    end

    def add_unresolved(source:, relation:, expression:, reason:, span: nil, attributes: nil)
      return unless source

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

    def connect_declaration(parent_id, child_id, span)
      add_edge(parent_id, child_id, "contains", span)
      add_edge(parent_id, child_id, "defines", span)
    end

    def resolve_superclasses
      resolve_semantics
    end
  end
end
