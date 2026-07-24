# frozen_string_literal: true

module LexiconRuby
  module RipperExtractor
    private

    CALL_NODES = %i[
      method_add_block method_add_arg command command_call call fcall vcall
      super zsuper yield yield0 aref
    ].freeze

    BRANCH_NODES = %i[if unless if_mod unless_mod case when rescue].freeze

    def visit(node, context)
      return unless node.is_a?(Array)
      return node.each { |child| visit(child, context) if child.is_a?(Array) } unless node[0].is_a?(Symbol)

      tag = node[0]
      if CALL_NODES.include?(tag)
        visit_call_expression(node, context)
        return
      end

      case tag
      when :program
        visit_body(node[1], context)
      when :module
        visit_module(node, context)
      when :class
        visit_class(node, context)
      when :sclass
        visit_singleton_class(node, context)
      when :def
        visit_method(node, context, singleton: context[:singleton])
      when :defs
        visit_method(node, context, singleton: true)
      when :assign
        visit_assignment(node, context)
      when :opassign
        visit_operator_assignment(node, context)
      when :massign
        visit_multiple_assignment(node, context)
      when :lambda
        visit_lambda(node, context)
      when :alias
        visit_alias(node, context)
      when :undef
        visit_undef(node, context)
      when :var_ref, :const_ref, :top_const_ref
        visit_variable_reference(node, context)
      when :binary
        operator = node[2].to_s
        register_operator_call(node, context, node[1], operator, [node[3]]) unless %w[&& || and or].include?(operator)
        visit(node[1], context)
        visit(node[3], context)
      when :unary
        register_operator_call(node, context, node[2], node[1].to_s, [])
        visit(node[2], context)
      else
        child_context = BRANCH_NODES.include?(tag) ? context.merge(branch_depth: context[:branch_depth] + 1) : context
        node.drop(1).each { |child| visit(child, child_context) if child.is_a?(Array) }
      end
    end

    def visit_children(node, context)
      node.drop(1).each { |child| visit(child, context) if child.is_a?(Array) }
    end

    def visit_body(body, context)
      return unless body

      if sexp_node?(body, :bodystmt)
        Array(body[1]).each { |part| visit(part, context) }
        body[2..].each { |part| visit(part, context) if part }
      elsif body.is_a?(Array) && body.first.is_a?(Array)
        body.each { |part| visit(part, context) }
      else
        visit(body, context)
      end
    end

    def visit_variable_reference(node, context)
      token = first_token(node)
      return unless token

      kind = case token[0]
             when :@ident then :local
             when :@ivar then :instance
             when :@cvar then :class_variable
             when :@const then :constant
             end
      return unless kind

      symbol = resolve_data_symbol(kind, token[1], context)
      add_dataflow_edge(context[:source_id], symbol, "reads", span_for(token)) if symbol
    end
  end
end
