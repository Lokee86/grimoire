# frozen_string_literal: true

module LexiconRuby
  module RipperExtractor
    private

    def visit(node, parent_id, namespace)
      return unless node.is_a?(Array)

      case node[0]
      when :program
        visit_body(node[1], parent_id, namespace)
      when :module
        visit_module(node, parent_id, namespace)
      when :class
        visit_class(node, parent_id, namespace)
      when :def
        visit_method(node, parent_id, namespace, singleton: false)
      when :defs
        visit_method(node, parent_id, namespace, singleton: true)
      when :assign
        visit_assignment(node, parent_id, namespace)
      when :command
        process_call(node[1], node[2], parent_id, namespace)
        visit(node[2], parent_id, namespace)
      when :command_call
        process_call(node[3], node[4], parent_id, namespace)
        visit(node[1], parent_id, namespace)
        visit(node[4], parent_id, namespace)
      when :method_add_arg
        call = node[1]
        process_call(call_name_token(call), node[2], parent_id, namespace)
        visit(call, parent_id, namespace)
        visit(node[2], parent_id, namespace)
      when :sclass
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: expression_for(node),
          reason: "unsupported-form",
          span: span_for(first_token(node))
        )
        visit_body(node[2], parent_id, namespace)
      when :alias, :undef
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: expression_for(node),
          reason: "unsupported-form",
          span: span_for(first_token(node))
        )
      else
        node.drop(1).each { |child| visit(child, parent_id, namespace) }
      end
    end

    def visit_body(body, parent_id, namespace)
      return unless body

      if sexp_node?(body, :bodystmt)
        body[1..].each { |part| visit(part, parent_id, namespace) }
      elsif body.is_a?(Array) && body.first.is_a?(Array)
        body.each { |part| visit(part, parent_id, namespace) }
      else
        visit(body, parent_id, namespace)
      end
    end

    def visit_module(node, parent_id, namespace)
      name = const_name(node[1])
      unless name
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: expression_for(node[1]),
          reason: "unsupported-form",
          span: span_for(first_token(node))
        )
        return
      end

      qualified_name = qualify(namespace, name)
      module_id = add_symbol("module", name.split("::").last, qualified_name, node[1], {})
      connect_declaration(parent_id, module_id, span_for(first_token(node[1])))
      visit_body(node[2], module_id, qualified_name)
    end

    def visit_class(node, parent_id, namespace)
      name = const_name(node[1])
      unless name
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: expression_for(node[1]),
          reason: "unsupported-form",
          span: span_for(first_token(node))
        )
        return
      end

      qualified_name = qualify(namespace, name)
      superclass = const_name(node[2])
      attributes = {}
      attributes["superclass"] = qualify(namespace, superclass) if superclass
      class_id = add_symbol("type", name.split("::").last, qualified_name, node[1], attributes)
      @type_nodes[qualified_name] << class_id
      connect_declaration(parent_id, class_id, span_for(first_token(node[1])))

      if node[2]
        if superclass
          @pending_extends << [class_id, qualify(namespace, superclass), span_for(first_token(node[2]))]
        else
          add_unresolved(
            source: class_id,
            relation: "extends",
            expression: expression_for(node[2]),
            reason: "unsupported-form",
            span: span_for(first_token(node[2]))
          )
        end
      end

      visit_body(node[3], class_id, qualified_name)
    end

    def visit_method(node, parent_id, namespace, singleton:)
      name_token = singleton ? node[2] : node[1]
      method_name = token_value(name_token)
      unless method_name
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: expression_for(node),
          reason: "unsupported-form",
          span: span_for(first_token(node))
        )
        return
      end

      owner_name = namespace.empty? ? "" : namespace
      qualified_name = if singleton
                        receiver = const_name(node[1])
                        receiver = qualify(namespace, receiver) if receiver
                        receiver ? "#{receiver}.#{method_name}" : "#{owner_name}.#{method_name}".delete_prefix(".")
                      else
                        owner_name.empty? ? method_name : "#{owner_name}##{method_name}"
                      end
      attributes = { "singleton" => singleton }
      parameters = parameter_names(singleton ? node[3] : node[2])
      attributes["parameters"] = parameters unless parameters.empty?
      method_id = add_symbol("method", method_name, qualified_name, name_token, attributes)
      connect_declaration(parent_id, method_id, span_for(name_token))
      visit_body(singleton ? node[4] : node[3], method_id, namespace)
    end

    def visit_assignment(node, parent_id, namespace)
      name = const_name(node[1])
      if name
        qualified_name = qualify(namespace, name)
        constant_id = add_symbol("constant", name.split("::").last, qualified_name, node[1], {})
        connect_declaration(parent_id, constant_id, span_for(first_token(node[1])))
      end
      visit(node[2], parent_id, namespace)
    end

    def process_call(name_token, arguments, parent_id, namespace)
      name = token_value(name_token)
      return unless name

      call_key = [@current_path, name, token_position(name_token)]
      return if @processed_calls[call_key]

      @processed_calls[call_key] = true
      if %w[require require_relative load].include?(name)
        process_import(name, name_token, arguments, parent_id)
      elsif LexiconRuby::Contract::METAPROGRAMMING_CALLS.include?(name)
        add_unresolved(
          source: parent_id,
          relation: "defines",
          expression: name,
          reason: "dynamic-target",
          span: span_for(name_token),
          attributes: { "namespace" => namespace }
        )
      end
    end

    def process_import(name, name_token, arguments, parent_id)
      required = literal_string(first_argument(arguments))
      if required
        import_id = add_symbol(
          "import",
          required,
          "#{@current_path}:#{required}:#{token_position(name_token).join(":")}",
          name_token,
          { "loader" => name, "target" => required },
          canonical_extra: "#{@current_path}\0#{token_position(name_token).join(":")}\0#{required}"
        )
        connect_declaration(parent_id, import_id, span_for(name_token))
        add_edge(parent_id, import_id, "imports", span_for(name_token))
      else
        add_unresolved(
          source: parent_id,
          relation: "imports",
          expression: expression_for(arguments),
          reason: "dynamic-target",
          span: span_for(name_token)
        )
      end
    end
  end
end
