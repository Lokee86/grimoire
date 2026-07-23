# frozen_string_literal: true

module LexiconRuby
  module CallEmission
    private

    def connect_constructor_initializers
      @constructors.each do |owner, constructor_id|
        initializers = @method_definitions.dig(owner, false, "initialize").to_a
        relation = initializers.length == 1 ? "calls" : "possible-calls"
        initializers.each { |initializer| add_edge(constructor_id, initializer, relation) }
      end
    end

    def emit_call_facts
      @call_sites.each_value do |call|
        targets = @resolved_targets.fetch(call.key, Set.new).to_a.sort
        if targets.length == 1
          add_edge(call.source, targets.first, "calls", call.span)
        elsif targets.length > 1
          targets.each { |target| add_edge(call.source, target, "possible-calls", call.span) }
          add_unresolved(
            source: call.source,
            relation: "calls",
            expression: call.expression,
            reason: multi_target_reason(call),
            span: call.span,
            attributes: call_attributes(call).merge("candidate_count" => targets.length)
          )
        else
          add_unresolved(
            source: call.source,
            relation: "calls",
            expression: call.expression,
            reason: unresolved_call_reason(call),
            span: call.span,
            attributes: call_attributes(call)
          )
        end
        add_edge(call.source, call.block_id, "possible-calls", call.span, { "block" => true }) if call.block_id
      end
    end

    def multi_target_reason(call)
      return "dynamic-target" if call.kind == :yield || call.receiver

      "ambiguous-target"
    end

    def call_attributes(call)
      attributes = { "name" => call.name, "kind" => call.kind.to_s }
      attributes["receiver"] = call.receiver_expression unless call.receiver_expression.to_s.empty?
      attributes
    end

    def unresolved_call_reason(call)
      return "dynamic-target" if call.kind == :yield
      return "builtin-target" if LexiconRuby::BUILTIN_CALLS.include?(call.name) || builtin_operator?(call) || literal_receiver?(call.receiver)
      return "external-target" if LexiconRuby::EXTERNAL_BARE_CALLS.include?(call.name)

      if call.receiver
        constant = const_name(call.receiver)
        return "external-target" if constant && resolve_constant_name(constant, call.namespace).nil?

        shape = expression_shape(call.receiver, call_context(call), before: call_position(call))
        return "external-target" if shape.singletons.any? { |name| external_dispatch?(name) }
        return "external-target" if shape.instances.any? { |name| external_dispatch?(name) }
        return "external-target" unless shape.elements.empty?
        return "dynamic-target"
      elsif external_dispatch?(call.owner) || external_source?(call.source)
        return "external-target"
      end
      "missing-target"
    end

    def builtin_operator?(call)
      %w[+ - * / % ** == != < > <= >= <=> === =~ !~ << >> & | ^ ~ ! -@ +@ [] []=].include?(call.name)
    end

    def literal_receiver?(receiver)
      receiver.is_a?(Array) && %i[
        array hash bare_assoc_hash string_literal symbol_literal regexp_literal
        @int @float @rational @imaginary
      ].include?(receiver[0])
    end

    def external_source?(source_id)
      path = @nodes[source_id]&.dig("path").to_s
      path.start_with?("test/", "db/", "config/")
    end
  end
end
