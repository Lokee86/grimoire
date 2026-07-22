# frozen_string_literal: true

module LexiconRuby
  module RipperAssignments
    private

    def visit_assignment(node, context)
      target = variable_target(node[1])
      value = node[2]
      if target&.dig(:kind) == :constant && factory_assignment?(value)
        visit_factory_assignment(node, target, context)
        return
      end

      if target&.dig(:kind) == :constant
        qualified_name = qualify(context[:namespace], target[:constant] || target[:name])
        constant_id = add_symbol("constant", qualified_name.split("::").last, qualified_name, node[1], {})
        connect_declaration(context[:parent_id], constant_id, span_for(first_token(node[1])))
        @constant_assignments[qualified_name] = value
      elsif target && %i[local instance class_variable].include?(target[:kind])
        @assignments << AssignmentInfo.new(
          scope: context[:scope_id],
          owner: context[:owner],
          singleton: context[:singleton],
          name: target[:name],
          kind: target[:kind],
          value: value,
          position: position_for(node),
          branch_dependent: context[:branch_depth].positive?
        )
      elsif sexp_node?(node[1], :field)
        receiver = node[1][1]
        name = "#{token_value(node[1][3])}="
        register_operator_call(node, context, receiver, name, [value])
        visit(receiver, context)
      elsif sexp_node?(node[1], :aref_field)
        arguments, keywords = argument_parts(node[1][2])
        register_manual_call(node, context, node[1][1], "[]=", arguments + [value], keywords, :receiver)
        visit(node[1][1], context)
        arguments.each { |argument| visit(argument, context) }
      end
      visit(value, context)
    end

    def visit_multiple_assignment(node, context)
      visit(node[2], context)
      collect_tokens(node[1]).each do |token|
        next unless %i[@ident @ivar].include?(token[0])

        @assignments << AssignmentInfo.new(
          scope: context[:scope_id],
          owner: context[:owner],
          singleton: context[:singleton],
          name: token[1],
          kind: token[0] == :@ivar ? :instance : :local,
          value: nil,
          position: token_position(token),
          branch_dependent: true
        )
      end
    end

    def visit_operator_assignment(node, context)
      target, operator, value = node[1], token_value(node[2]).to_s, node[3]
      receiver = if sexp_node?(target, :field) || sexp_node?(target, :aref_field)
                   target[1]
                 else
                   target
                 end
      register_operator_call(node, context, receiver, operator.delete_suffix("="), [value])
      visit(target, context)
      visit(value, context)
    end

    def visit_factory_assignment(node, target, context)
      value = node[2]
      info = decompose_call(value)
      qualified_name = qualify(context[:namespace], target[:constant] || target[:name])
      factory = const_name(info[:receiver]).to_s.delete_prefix("::")
      kind = factory == "Module" ? "module" : "type"
      attributes = { "factory" => "#{factory}.#{info[:name]}" }
      type_id = add_symbol(kind, qualified_name.split("::").last, qualified_name, node[1], attributes)
      register_type(qualified_name, kind, type_id, context[:namespace])
      connect_declaration(context[:parent_id], type_id, span_for(first_token(node[1])))
      ensure_constructor(qualified_name) if kind == "type"
      fields = info[:arguments].filter_map { |argument| literal_symbol(argument) }
      if %w[Struct Data].include?(factory)
        fields.each do |field|
          add_synthetic_method(qualified_name, field, singleton: false, attribute: true)
          add_synthetic_method(qualified_name, "#{field}=", singleton: false, attribute: true) if factory == "Struct"
        end
      end
      @constant_type_assignments[qualified_name] = qualified_name
      register_call(info.merge(block: nil), context, nil)
      info[:arguments].each { |argument| visit(argument, context) }
      if info[:block]
        visit_body(
          block_parts(info[:block])[1],
          declaration_context(context, type_id, qualified_name, singleton: false)
        )
      end
    end

    def factory_assignment?(value)
      info = decompose_call(value)
      return false unless info && info[:receiver]

      factory = const_name(info[:receiver]).to_s.delete_prefix("::")
      (info[:name] == "new" && %w[Struct Class Module].include?(factory)) ||
        (info[:name] == "define" && factory == "Data")
    end

    def visit_alias(node, context)
      return unresolved_declaration(node, context) unless context[:owner]

      new_name = literal_symbol(node[1])
      old_name = literal_symbol(node[2])
      return unresolved_declaration(node, context) unless new_name && old_name

      original_ids = @method_definitions.dig(context[:owner], context[:singleton], old_name).to_a
      alias_id = add_synthetic_method(
        context[:owner],
        new_name,
        singleton: context[:singleton],
        alias_of: original_ids.first
      )
      @aliases[[context[:owner], context[:singleton], new_name]] = old_name
      add_edge(alias_id, original_ids.first, "calls", span_for_expression(node)) if original_ids.length == 1
    end

    def visit_undef(node, context)
      return unless context[:owner]

      node.drop(1).filter_map { |value| literal_symbol(value) }.each do |name|
        @undefined_methods[[context[:owner], context[:singleton]]] << name
      end
    end

    def register_operator_call(node, context, receiver, name, arguments)
      register_manual_call(node, context, receiver, name, arguments, {}, :receiver)
    end

    def register_manual_call(node, context, receiver, name, arguments, keywords, kind)
      register_call(
        {
          node: node,
          receiver: receiver,
          name_token: first_token(node),
          name: name,
          arguments: arguments,
          keywords: keywords,
          kind: kind,
          block: nil
        },
        context,
        nil
      )
    end
  end
end
