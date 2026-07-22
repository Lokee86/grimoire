# frozen_string_literal: true

module LexiconRuby
  module CallShapes
    private

    ACTIVE_RECORD_INSTANCE_METHODS = %w[
      find find_by find_by! create create! first first! last last! take take!
      find_or_create_by find_or_create_by! find_or_initialize_by new build
    ].freeze
    ACTIVE_RECORD_COLLECTION_METHODS = %w[
      all where order reorder limit offset joins includes preload eager_load
      left_joins left_outer_joins distinct select group having
    ].freeze

    def resolve_constant_name(reference, namespace)
      return nil if reference.nil? || reference.empty?

      lexical_constant_candidates(namespace, reference).find do |candidate|
        @types.key?(candidate) || @constant_type_assignments.key?(candidate)
      end
    end

    def expression_shape(expression, context, before: nil, seen: Set.new)
      return EMPTY_SHAPE unless expression.is_a?(Array)

      marker = [context[:scope_id], expression.object_id]
      return EMPTY_SHAPE if seen.include?(marker)

      next_seen = seen | [marker]
      tag = expression[0]
      if token?(expression)
        return token_shape(expression, context, before, next_seen)
      end

      case tag
      when :var_ref, :var_field
        token_shape(expression[1], context, before, next_seen)
      when :const_ref, :top_const_ref, :const_path_ref
        constant_shape(const_name(expression), context[:namespace])
      when :method_add_block, :method_add_arg, :command, :command_call, :call, :fcall, :vcall, :super, :zsuper, :yield, :yield0
        call_result_shape(expression, context, before, next_seen)
      when :aref
        result = call_result_shape(expression, context, before, next_seen)
        return result unless result.empty?

        expression_shape(expression[1], context, before: before, seen: next_seen).element_shape
      when :lambda
        lambda_id = @lambda_ids[expression.object_id]
        lambda_id ? ValueShape.new(callables: [lambda_id]) : EMPTY_SHAPE
      when :array
        element_shape(Array(expression[1]), context, before, next_seen)
      when :hash, :bare_assoc_hash
        values = association_list(expression).map { |association| association[2] }
        element_shape(values, context, before, next_seen)
      when :ifop
        expression_shape(expression[2], context, before: before, seen: next_seen).merge(
          expression_shape(expression[3], context, before: before, seen: next_seen)
        )
      when :paren
        expression_shape(expression[1], context, before: before, seen: next_seen)
      when :binary
        call_result_shape(expression, context, before, next_seen)
      else
        EMPTY_SHAPE
      end
    end

    def token_shape(token, context, before, seen)
      return EMPTY_SHAPE unless token?(token)

      case token[0]
      when :@const
        constant_shape(token[1], context[:namespace])
      when :@ident
        local_shape(token[1], context, before, seen)
      when :@ivar, :@cvar
        assignment_shape(token[1], token[0] == :@ivar ? :instance : :class_variable, context, before, seen)
      when :@kw
        if token[1] == "self" && context[:owner]
          context[:singleton] ? ValueShape.new(singletons: [context[:owner]]) : ValueShape.new(instances: [context[:owner]])
        else
          EMPTY_SHAPE
        end
      else
        EMPTY_SHAPE
      end
    end

    def constant_shape(reference, namespace)
      resolved = resolve_constant_name(reference, namespace)
      resolved ? ValueShape.new(singletons: [resolved]) : EMPTY_SHAPE
    end

    def local_shape(name, context, before, seen)
      scope = context[:scope_id]
      return EMPTY_SHAPE unless scope

      marker = [scope, name, :local]
      return EMPTY_SHAPE if seen.include?(marker)

      next_seen = seen | [marker]
      shape = @parameter_shapes.fetch([scope, name], EMPTY_SHAPE)
      method = @methods[scope]
      if method
        default = method.parameters[:optional].to_h[name] || method.parameters[:keywords][name]
        shape = shape.merge(expression_shape(default, context, before: before, seen: next_seen)) if default && default != false
      end
      relevant = @assignments.select do |assignment|
        assignment.scope == scope && assignment.kind == :local && assignment.name == name &&
          (before.nil? || (assignment.position <=> before) <= 0)
      end.sort_by(&:position)
      relevant.each do |assignment|
        candidate = expression_shape(
          assignment.value,
          context,
          before: assignment.position,
          seen: next_seen
        )
        shape = assignment.branch_dependent ? shape.merge(candidate) : candidate
      end
      return shape unless shape.empty?

      parent_scope = @scope_parents[scope]
      if parent_scope
        parent_method = @methods[parent_scope]
        parent_context = context.merge(
          scope_id: parent_scope,
          method_id: parent_scope,
          owner: parent_method&.owner || context[:owner],
          singleton: parent_method&.singleton || context[:singleton]
        )
        return local_shape(name, parent_context, before, next_seen)
      end
      EMPTY_SHAPE
    end

    def assignment_shape(name, kind, context, before, seen)
      marker = [context[:owner], context[:singleton], name, kind]
      return EMPTY_SHAPE if seen.include?(marker)

      next_seen = seen | [marker]
      shape = EMPTY_SHAPE
      relevant = @assignments.select do |assignment|
        assignment.kind == kind && assignment.name == name && assignment.owner == context[:owner] &&
          assignment.singleton == context[:singleton]
      end.sort_by(&:position)
      relevant.each do |assignment|
        assignment_context = context.merge(scope_id: assignment.scope)
        candidate = expression_shape(
          assignment.value,
          assignment_context,
          before: assignment.position,
          seen: next_seen
        )
        shape = assignment.branch_dependent ? shape.merge(candidate) : candidate
      end
      shape
    end

    def element_shape(expressions, context, before, seen)
      instances = expressions.flat_map do |item|
        expression_shape(item, context, before: before, seen: seen).instances.to_a
      end
      ValueShape.new(elements: instances)
    end

    def call_result_shape(expression, context, before, seen)
      call = call_site_for_expression(expression, context[:source_id])
      info = decompose_call(expression)
      return EMPTY_SHAPE unless info

      receiver_shape = info[:receiver] && expression_shape(info[:receiver], context, before: before, seen: seen)
      if info[:name] == "new" && receiver_shape && !receiver_shape.singletons.empty?
        return ValueShape.new(instances: receiver_shape.singletons)
      end
      if info[:name] == "class" && self_reference?(info[:receiver]) && context[:owner]
        return ValueShape.new(singletons: [context[:owner]])
      end
      if receiver_shape && ACTIVE_RECORD_INSTANCE_METHODS.include?(info[:name])
        local_types = receiver_shape.singletons.select { |name| active_record_type?(name) }
        return ValueShape.new(instances: local_types) unless local_types.empty?
      end
      if receiver_shape && ACTIVE_RECORD_COLLECTION_METHODS.include?(info[:name])
        local_types = receiver_shape.singletons.select { |name| active_record_type?(name) }
        return ValueShape.new(elements: local_types) unless local_types.empty?
      end

      targets = call ? @resolved_targets.fetch(call.key, Set.new) : Set.new
      targets.reduce(EMPTY_SHAPE) do |result, target_id|
        method = @methods[target_id]
        target_shape = if @nodes[target_id]&.dig("kind") == "constructor" && method&.owner
                         ValueShape.new(instances: [method.owner])
                       else
                         @method_returns.fetch(target_id, EMPTY_SHAPE)
                       end
        result.merge(target_shape)
      end
    end

    def active_record_type?(name, seen = Set.new)
      return true if name == "ActiveRecord::Base"
      return false if name.nil? || seen.include?(name)

      info = type_info(name)
      return false unless info

      active_record_type?(info.superclass_ref, seen | [name])
    end

    def call_site_for_expression(expression, source_id)
      key = @call_node_keys[[source_id, expression.object_id]]
      key && @call_sites[key]
    end

    def method_context(method)
      {
        source_id: method.id,
        scope_id: method.id,
        method_id: method.id,
        owner: method.owner,
        singleton: method.singleton,
        namespace: method.namespace
      }
    end

    def infer_return_shape(method)
      alias_target = @method_alias_targets[method.id]
      return @method_returns.fetch(alias_target, EMPTY_SHAPE) if alias_target
      return ValueShape.new(instances: [method.owner]) if @nodes[method.id]&.dig("kind") == "constructor"
      return EMPTY_SHAPE unless method.body

      context = method_context(method)
      expressions = return_expressions(method.body)
      expressions.reduce(EMPTY_SHAPE) do |shape, expression|
        shape.merge(expression_shape(expression, context, before: position_for(expression)))
      end
    end

    def return_expressions(body)
      explicit = []
      collect_explicit_returns(body, explicit)
      terminals = terminal_expressions(body)
      (explicit + terminals).compact.uniq(&:object_id)
    end

    def collect_explicit_returns(node, result)
      return unless node.is_a?(Array)
      return if %i[def defs class module sclass lambda].include?(node[0])

      if node[0] == :return
        positional, = argument_parts(node[1])
        result.concat(positional)
        return
      end
      node.drop(1).each { |child| collect_explicit_returns(child, result) if child.is_a?(Array) }
    end

    def terminal_expressions(node)
      return [] unless node.is_a?(Array)
      return [node] if token?(node)

      case node[0]
      when :bodystmt
        statements = Array(node[1])
        values = statements.empty? ? [] : terminal_expressions(statements.last)
        values + terminal_expressions(node[2]) + terminal_expressions(node[3])
      when :if, :unless
        terminal_expressions(node[2]) + terminal_expressions(node[3])
      when :if_mod, :unless_mod
        terminal_expressions(node[1]) + terminal_expressions(node[2])
      when :else, :elsif, :when, :rescue, :ensure, :begin
        node.drop(1).flat_map { |child| terminal_expressions(child) }
      when :return, :return0, :void_stmt
        []
      when :def, :defs, :class, :module, :sclass
        []
      else
        [node]
      end
    end
  end
end
