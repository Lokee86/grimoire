# frozen_string_literal: true

module LexiconRuby
  module RipperCalls
    private

    STRUCTURAL_COMMANDS = %w[
      include prepend extend attr_reader attr_writer attr_accessor
      private protected public module_function private_class_method public_class_method
    ].freeze

    def visit_call_expression(node, context)
      info = decompose_call(node)
      return visit_children(node, context) unless info

      name = info[:name]
      if structural_call?(name, context)
        process_structural_call(name, info, context)
      elsif explicit_structural_call?(name, info)
        process_explicit_structural_call(name, info, context)
      elsif %w[included class_methods].include?(name) && info[:block]
        visit_hook_block(name, info[:block], context)
      else
        block_id = info[:block] && create_block(info[:block], context, info)
        register_call(info, context, block_id)
      end

      visit(info[:receiver], context) if info[:receiver]
      info[:arguments].each { |argument| visit(argument, context) }
      info[:keywords].each_value { |argument| visit(argument, context) }
    end

    def decompose_call(node)
      return nil unless node.is_a?(Array)

      case node[0]
      when :method_add_block
        inner = decompose_call(node[1])
        inner&.merge(node: node, block: node[2])
      when :method_add_arg
        inner = decompose_call(node[1])
        return nil unless inner

        positional, keywords = argument_parts(node[2])
        inner.merge(node: node, arguments: positional, keywords: keywords)
      when :command
        positional, keywords = argument_parts(node[2])
        call_hash(node, nil, node[1], positional, keywords, :bare)
      when :command_call
        positional, keywords = argument_parts(node[4])
        call_hash(node, node[1], node[3], positional, keywords, :receiver)
      when :call
        call_hash(node, node[1], node[3], [], {}, :receiver)
      when :fcall, :vcall
        call_hash(node, nil, node[1], [], {}, :bare)
      when :super
        positional, keywords = argument_parts(node[1])
        call_hash(node, nil, nil, positional, keywords, :super, name: "super")
      when :zsuper
        call_hash(node, nil, nil, [], {}, :super, name: "super")
      when :yield
        positional, keywords = argument_parts(node[1])
        call_hash(node, nil, nil, positional, keywords, :yield, name: "yield")
      when :yield0
        call_hash(node, nil, nil, [], {}, :yield, name: "yield")
      when :aref
        positional, keywords = argument_parts(node[2])
        call_hash(node, node[1], nil, positional, keywords, :receiver, name: "[]")
      end
    end

    def call_hash(node, receiver, name_token, positional, keywords, kind, name: nil)
      call_name = name || token_value(name_token)
      return nil unless call_name

      {
        node: node,
        receiver: receiver,
        name_token: name_token || first_token(node),
        name: call_name,
        arguments: positional,
        keywords: keywords,
        kind: kind,
        block: nil
      }
    end

    def register_call(info, context, block_id)
      name = info[:name]
      if %w[require require_relative load].include?(name) && info[:kind] == :bare
        process_import(name, info[:name_token], info[:arguments].first, info[:node], context[:source_id])
        return
      end

      if LexiconRuby::Contract::METAPROGRAMMING_CALLS.include?(name)
        add_unresolved(
          source: context[:source_id],
          relation: "defines",
          expression: source_expression(info[:node]),
          reason: "dynamic-target",
          span: span_for_expression(info[:node]),
          attributes: { "namespace" => context[:namespace] }
        )
        return
      end

      span = span_for_expression(info[:node])
      key = [@current_path, context[:source_id], name, span&.values_at("start_line", "start_column", "end_line", "end_column")]
      return if @call_sites.key?(key)

      @call_sites[key] = CallSite.new(
        key: key,
        source: context[:source_id],
        owner: context[:owner],
        singleton: context[:singleton],
        namespace: context[:namespace],
        receiver: info[:receiver],
        receiver_expression: info[:receiver] ? source_expression(info[:receiver]) : nil,
        name: name,
        arguments: info[:arguments],
        keywords: info[:keywords],
        block_id: block_id,
        expression: source_expression(info[:node]),
        span: span,
        node: info[:node],
        kind: info[:kind]
      )
    end

    def structural_call?(name, context)
      context[:owner] && context[:method_id].nil? && STRUCTURAL_COMMANDS.include?(name)
    end

    def explicit_structural_call?(name, info)
      %w[include prepend extend].include?(name) && const_name(info[:receiver]) &&
        info[:arguments].any? { |argument| const_name(argument) }
    end

    def process_explicit_structural_call(name, info, context)
      receiver_reference = const_name(info[:receiver])
      owner = resolve_constant_name(receiver_reference, context[:namespace])
      owner ||= external_type(receiver_reference.to_s.delete_prefix("::"), kind: "type")
      source_id = type_info(owner)&.ids&.first || context[:source_id]
      references = info[:arguments].filter_map { |argument| const_name(argument) }
      @pending_mixin_edges << [
        owner, source_id, name, references,
        span_for_expression(info[:node]), context[:namespace]
      ]
    end

    def process_structural_call(name, info, context)
      case name
      when "include", "prepend", "extend"
        references = info[:arguments].filter_map { |argument| const_name(argument) }
        collection = case name
                     when "include" then type_info(context[:owner]).includes
                     when "prepend" then type_info(context[:owner]).prepends
                     else type_info(context[:owner]).extends_modules
                     end
        references.each { |reference| collection << reference unless collection.include?(reference) }
        @pending_mixin_edges << [
          context[:owner], context[:source_id], name, references,
          span_for_expression(info[:node]), context[:namespace]
        ]
      when "attr_reader", "attr_writer", "attr_accessor"
        info[:arguments].filter_map { |argument| literal_symbol(argument) }.each do |attribute|
          add_attribute_methods(context, attribute, name)
        end
      when "module_function"
        names = info[:arguments].filter_map { |argument| literal_symbol(argument) }
        if names.empty?
          @module_function_all << context[:owner]
        else
          names.each { |method_name| copy_module_function(context[:owner], method_name) }
        end
      when "private_class_method", "public_class_method"
        nil
      end
    end

    def add_attribute_methods(context, attribute, command)
      if %w[attr_reader attr_accessor].include?(command)
        add_synthetic_method(context[:owner], attribute, singleton: context[:singleton], attribute: true)
      end
      if %w[attr_writer attr_accessor].include?(command)
        add_synthetic_method(context[:owner], "#{attribute}=", singleton: context[:singleton], attribute: true)
      end
    end

    def copy_module_function(owner, method_name)
      @method_definitions.dig(owner, false, method_name).to_a.each do |method_id|
        original = @methods[method_id]
        add_synthetic_method(owner, method_name, singleton: true, alias_of: method_id, body: original&.body)
      end
    end

    def process_import(name, name_token, argument, expression, parent_id)
      required = literal_string(argument)
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
        record_local_dependency(@current_path, required) if name == "require_relative"
      else
        add_unresolved(
          source: parent_id,
          relation: "imports",
          expression: source_expression(expression),
          reason: "dynamic-target",
          span: span_for(name_token)
        )
      end
    end
  end
end
