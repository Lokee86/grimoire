# frozen_string_literal: true

module LexiconRuby
  module RipperDeclarations
    private

    def visit_module(node, context)
      name = const_name(node[1])
      return unresolved_declaration(node, context) unless name

      qualified_name = qualify(context[:namespace], name)
      module_id = add_symbol("module", qualified_name.split("::").last, qualified_name, node[1], {})
      register_type(qualified_name, "module", module_id, context[:namespace])
      connect_declaration(context[:parent_id], module_id, span_for(first_token(node[1])))
      visit_body(node[2], declaration_context(context, module_id, qualified_name, singleton: false))
    end

    def visit_class(node, context)
      name = const_name(node[1])
      return unresolved_declaration(node, context) unless name

      qualified_name = qualify(context[:namespace], name)
      superclass = const_name(node[2])
      attributes = {}
      attributes["superclass"] = superclass if superclass
      class_id = add_symbol("type", qualified_name.split("::").last, qualified_name, node[1], attributes)
      info = register_type(qualified_name, "type", class_id, context[:namespace])
      info.superclass_ref ||= superclass
      connect_declaration(context[:parent_id], class_id, span_for(first_token(node[1])))
      ensure_constructor(qualified_name)
      @pending_extends << [class_id, qualified_name, superclass, span_for(first_token(node[2]))] if node[2]
      visit_body(node[3], declaration_context(context, class_id, qualified_name, singleton: false))
    end

    def visit_singleton_class(node, context)
      owner = if self_reference?(node[1])
                context[:owner]
              else
                reference = const_name(node[1])
                resolve_constant_name(reference, context[:namespace]) || qualify(context[:namespace], reference)
              end
      unless owner
        add_unresolved(
          source: context[:source_id],
          relation: "defines",
          expression: source_expression(node[1]),
          reason: "unsupported-form",
          span: span_for_expression(node[1])
        )
        return
      end

      parent_id = type_info(owner)&.ids&.first || context[:parent_id]
      visit_body(
        node[2],
        context.merge(
          parent_id: parent_id,
          source_id: parent_id,
          namespace: owner,
          owner: owner,
          singleton: true,
          method_id: nil,
          scope_id: parent_id
        )
      )
    end

    def visit_method(node, context, singleton:)
      name_token = node[0] == :defs ? node[3] : node[1]
      method_name = token_value(name_token)
      return unresolved_declaration(node, context) unless method_name

      owner = context[:owner].to_s
      if node[0] == :defs && !self_reference?(node[1])
        receiver = const_name(node[1])
        owner = resolve_constant_name(receiver, context[:namespace]) || qualify(context[:namespace], receiver)
      end
      separator = singleton ? "." : "#"
      qualified_name = owner.empty? ? method_name : "#{owner}#{separator}#{method_name}"
      params = node[0] == :defs ? node[4] : node[2]
      body = node[0] == :defs ? node[5] : node[3]
      spec = parameter_spec(params)
      attributes = { "singleton" => singleton }
      names = parameter_names(params)
      attributes["parameters"] = names unless names.empty?
      method_id = add_symbol("method", method_name, qualified_name, name_token, attributes)
      parent_id = type_info(owner)&.ids&.first || context[:parent_id]
      connect_declaration(parent_id, method_id, span_for(name_token))
      register_method(method_id, owner, method_name, singleton, spec, body, context[:namespace], token_position(name_token))
      if !singleton && @module_function_all.include?(owner)
        add_synthetic_method(owner, method_name, singleton: true, alias_of: method_id, body: body)
      end
      method_context = context.merge(
        parent_id: method_id,
        source_id: method_id,
        owner: owner,
        singleton: singleton,
        method_id: method_id,
        scope_id: method_id,
        branch_depth: 0
      )
      visit_parameter_defaults(spec, method_context)
      visit_body(body, method_context)
    end

    def visit_parameter_defaults(spec, context)
      spec[:optional].each { |_name, value| visit(value, context) }
      spec[:keywords].each_value { |value| visit(value, context) unless value == false }
    end

    def unresolved_declaration(node, context)
      add_unresolved(
        source: context[:source_id],
        relation: "defines",
        expression: source_expression(node),
        reason: "unsupported-form",
        span: span_for_expression(node)
      )
    end

    def declaration_context(context, node_id, qualified_name, singleton:)
      context.merge(
        parent_id: node_id,
        source_id: node_id,
        namespace: qualified_name,
        owner: qualified_name,
        singleton: singleton,
        method_id: nil,
        scope_id: node_id,
        branch_depth: 0
      )
    end

    def register_type(name, kind, id, declaration_namespace, external: false)
      info = @types[name] ||= TypeInfo.new(
        name: name,
        kind: kind,
        ids: [],
        declaration_namespace: declaration_namespace,
        superclass_ref: nil,
        includes: [],
        prepends: [],
        extends_modules: [],
        external: external
      )
      info.ids << id unless info.ids.include?(id)
      info
    end

    def type_info(name)
      @types[name]
    end

    def register_method(id, owner, name, singleton, parameters, body, namespace, position, synthetic: false)
      @methods[id] = MethodInfo.new(
        id: id,
        owner: owner,
        name: name,
        singleton: singleton,
        parameters: parameters,
        body: body,
        namespace: namespace,
        path: @current_path,
        position: position,
        synthetic: synthetic
      )
      @method_definitions[owner][singleton][name] << id
      @method_contexts[id] = { owner: owner, singleton: singleton }
      id
    end

    def add_synthetic_method(owner, name, singleton:, attribute: false, alias_of: nil, body: nil)
      existing = @method_definitions.dig(owner, singleton, name).to_a.find do |method_id|
        @methods[method_id]&.synthetic
      end
      return existing if existing

      parent_id = type_info(owner)&.ids&.first
      return nil unless parent_id

      separator = singleton ? "." : "#"
      qualified_name = "#{owner}#{separator}#{name}"
      kind = name == "new" && singleton ? "constructor" : "method"
      attributes = { "singleton" => singleton, "synthetic" => true }
      attributes["attribute"] = true if attribute
      attributes["alias_of"] = alias_of if alias_of
      method_id = add_node(
        kind: kind,
        name: name,
        path: @nodes[parent_id]["path"],
        qualified_name: qualified_name,
        canonical: "synthetic\0#{qualified_name}",
        attributes: attributes
      )
      connect_declaration(parent_id, method_id, nil)
      @method_alias_targets[method_id] = alias_of if alias_of
      register_method(
        method_id,
        owner,
        name,
        singleton,
        { positional: [], optional: [], keywords: {}, rest: nil, keyword_rest: nil, block: nil },
        body,
        owner,
        [0, 0],
        synthetic: true
      )
    end

    def ensure_constructor(owner)
      @constructors[owner] ||= add_synthetic_method(owner, "new", singleton: true)
    end
  end
end
