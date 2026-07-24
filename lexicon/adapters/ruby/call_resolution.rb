# frozen_string_literal: true

module LexiconRuby
  module CallResolution
    private

    def resolve_semantics
      resolve_hierarchy
      index_call_sites
      16.times do
        changed = resolve_call_iteration
        break unless changed
      end
      connect_constructor_initializers
      emit_call_facts
    end

    def resolve_call_targets(call)
      return resolve_super_targets(call) if call.kind == :super
      return @passed_blocks.fetch(call.source, Set.new).dup if call.kind == :yield

      context = call_context(call)
      return Set.new(lookup_bare(call.owner, call.singleton, call.name, call.source)) if call.receiver.nil?

      shape = expression_shape(call.receiver, context, before: call_position(call))
      targets = Set.new
      targets.merge(shape.callables) if call.name == "call"
      shape.singletons.each do |type_name|
        if call.name == "new"
          constructor = @constructors[type_name]
          targets << constructor if constructor
        else
          targets.merge(lookup_singleton(type_name, call.name))
        end
      end
      shape.instances.each { |type_name| targets.merge(lookup_instance(type_name, call.name)) }
      targets
    end

    def call_context(call)
      method = @methods[call.source]
      {
        source_id: call.source,
        scope_id: call.source,
        method_id: call.source,
        owner: method&.owner || call.owner,
        singleton: method.nil? ? call.singleton : method.singleton,
        namespace: method&.namespace || call.namespace
      }
    end

    def call_position(call)
      span = call.span || {}
      [span.fetch("start_line", 0), span.fetch("start_column", 0)]
    end
  end
end
