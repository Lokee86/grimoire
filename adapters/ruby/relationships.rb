# frozen_string_literal: true

module LexiconRuby
  module Relationships
    private

    def add_edge(source, target, relation, span = nil)
      record = { "record" => "edge", "source" => source, "target" => target, "relation" => relation }
      record["span"] = span if span
      @edges[[source, target, relation]] ||= record
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

    def connect_declaration(parent_id, child_id, span)
      add_edge(parent_id, child_id, "contains", span)
      add_edge(parent_id, child_id, "defines", span)
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

      resolve_local_calls
    end

    def resolve_local_calls
      @pending_calls.each do |call|
        method_ids = @method_definitions.dig(call[:owner], call[:name])
        if method_ids.length == 1
          add_edge(call[:source], method_ids.first, "calls", call[:span])
          next
        end

        reason = method_ids.empty? ? "missing-target" : "ambiguous-target"
        add_unresolved(
          source: call[:source],
          relation: "calls",
          expression: call[:expression],
          reason: reason,
          span: call[:span]
        )
      end
    end
  end
end
