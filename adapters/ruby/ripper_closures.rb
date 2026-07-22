# frozen_string_literal: true

module LexiconRuby
  module RipperClosures
    private

    def visit_hook_block(name, block, context)
      hook_context = context.merge(singleton: name == "class_methods", mixin_hook: name)
      visit_block_body(block, hook_context)
    end

    def create_block(block, context, call_info)
      token = first_token(block)
      line, column = token_position(token)
      qualified_name = "#{context[:owner] || @current_path}.<block>@#{line}:#{column + 1}"
      block_id = add_symbol(
        "function",
        "<block>@#{line}:#{column + 1}",
        qualified_name,
        token,
        { "block" => true },
        canonical_extra: context[:source_id]
      )
      connect_declaration(context[:source_id], block_id, span_for(token))
      params, body = block_parts(block)
      spec = parameter_spec(params)
      @methods[block_id] = MethodInfo.new(
        id: block_id,
        owner: context[:owner],
        name: qualified_name,
        singleton: context[:singleton],
        parameters: spec,
        body: body,
        namespace: context[:namespace],
        path: @current_path,
        position: token_position(token),
        synthetic: true
      )
      @blocks[block_id] = BlockInfo.new(
        id: block_id,
        owner: context[:owner],
        singleton: context[:singleton],
        parameters: spec,
        call_key: call_info[:node].object_id,
        body: body,
        path: @current_path
      )
      visit_body(
        body,
        context.merge(
          parent_id: block_id,
          source_id: block_id,
          scope_id: block_id,
          method_id: block_id
        )
      )
      block_id
    end

    def visit_block_body(block, context)
      _params, body = block_parts(block)
      visit_body(body, context)
    end

    def block_parts(block)
      return [nil, nil] unless block.is_a?(Array)

      case block[0]
      when :do_block, :brace_block
        params = block[1]
        params = params[1] if sexp_node?(params, :block_var)
        [params, block[2]]
      else
        [nil, block]
      end
    end

    def visit_lambda(node, context)
      token = first_token(node)
      line, column = token_position(token)
      qualified_name = "#{context[:owner] || @current_path}.<lambda>@#{line}:#{column + 1}"
      lambda_id = add_symbol(
        "function",
        "<lambda>@#{line}:#{column + 1}",
        qualified_name,
        token,
        { "lambda" => true },
        canonical_extra: context[:source_id]
      )
      connect_declaration(context[:source_id], lambda_id, span_for(token))
      params = node[1]
      body = node[2]
      spec = parameter_spec(params)
      @methods[lambda_id] = MethodInfo.new(
        id: lambda_id,
        owner: context[:owner],
        name: qualified_name,
        singleton: context[:singleton],
        parameters: spec,
        body: body,
        namespace: context[:namespace],
        path: @current_path,
        position: token_position(token),
        synthetic: true
      )
      @lambda_ids[node.object_id] = lambda_id
      visit_body(
        body,
        context.merge(
          parent_id: lambda_id,
          source_id: lambda_id,
          scope_id: lambda_id,
          method_id: lambda_id
        )
      )
    end
  end
end
