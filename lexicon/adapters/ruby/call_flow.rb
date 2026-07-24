# frozen_string_literal: true

module LexiconRuby
  module CallFlow
    private

    def index_call_sites
      @call_sites.each_value do |call|
        @call_node_keys[[call.source, call.node.object_id]] = call.key
      end
    end

    def resolve_call_iteration
      changed = false
      @call_sites.each_value do |call|
        targets = resolve_call_targets(call)
        previous = @resolved_targets.fetch(call.key, Set.new)
        if targets != previous
          @resolved_targets[call.key] = targets
          changed = true
        end
      end
      rebuild_passed_blocks
      changed = propagate_arguments || changed
      changed = propagate_block_parameters || changed
      @methods.each_value do |method|
        inferred = infer_return_shape(method)
        previous = @method_returns.fetch(method.id, EMPTY_SHAPE)
        next if inferred == previous

        @method_returns[method.id] = inferred
        changed = true
      end
      changed
    end

    def rebuild_passed_blocks
      @passed_blocks.clear
      @call_sites.each_value do |call|
        next unless call.block_id

        @resolved_targets.fetch(call.key, Set.new).each { |target| @passed_blocks[target] << call.block_id }
      end
    end

    def propagate_arguments
      changed = false
      @call_sites.each_value do |call|
        context = call_context(call)
        @resolved_targets.fetch(call.key, Set.new).each do |target_id|
          method = @methods[target_id]
          next unless method

          method = initializer_for_constructor(method) || method
          positional_names = method.parameters[:positional] + method.parameters[:optional].map(&:first)
          positional_names.each_with_index do |name, index|
            argument = call.arguments[index]
            next unless argument

            shape = expression_shape(argument, context, before: call_position(call))
            changed = merge_parameter_shape(method.id, name, shape) || changed
          end
          method.parameters[:keywords].each_key do |name|
            argument = call.keywords[name]
            next unless argument

            shape = expression_shape(argument, context, before: call_position(call))
            changed = merge_parameter_shape(method.id, name, shape) || changed
          end
          if call.block_id && method.parameters[:block]
            changed = merge_parameter_shape(
              method.id,
              method.parameters[:block],
              ValueShape.new(callables: [call.block_id])
            ) || changed
          end
        end
      end
      changed
    end

    def initializer_for_constructor(method)
      return nil unless @nodes[method.id]&.dig("kind") == "constructor"

      ids = @method_definitions.dig(method.owner, false, "initialize").to_a
      ids.length == 1 ? @methods[ids.first] : nil
    end

    def merge_parameter_shape(method_id, name, shape)
      return false if name.nil? || shape.empty?

      key = [method_id, name]
      previous = @parameter_shapes.fetch(key, EMPTY_SHAPE)
      merged = previous.merge(shape)
      return false if merged == previous

      @parameter_shapes[key] = merged
      true
    end

    def propagate_block_parameters
      changed = false
      @call_sites.each_value do |call|
        next unless call.block_id

        block = @blocks[call.block_id]
        first_name = block&.parameters&.dig(:positional)&.first
        next unless first_name && call.receiver

        receiver_shape = expression_shape(call.receiver, call_context(call), before: call_position(call))
        changed = merge_parameter_shape(call.block_id, first_name, receiver_shape.element_shape) || changed
      end
      @call_sites.each_value do |yield_call|
        next unless yield_call.kind == :yield

        @passed_blocks.fetch(yield_call.source, Set.new).each do |block_id|
          block = @blocks[block_id]
          next unless block

          block.parameters[:positional].each_with_index do |name, index|
            argument = yield_call.arguments[index]
            next unless argument

            shape = expression_shape(argument, call_context(yield_call), before: call_position(yield_call))
            changed = merge_parameter_shape(block_id, name, shape) || changed
          end
        end
      end
      changed
    end
  end
end
