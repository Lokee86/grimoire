# frozen_string_literal: true

module LexiconRuby
  module RipperSyntax
    private

    def const_name(value)
      return nil unless value
      return value[1] if token?(value) && value[0] == :@const

      if value.is_a?(Array)
        case value[0]
        when :const_ref, :var_ref, :var_field, :top_const_ref
          name = const_name(value[1])
          return name && (value[0] == :top_const_ref ? "::#{name}" : name)
        when :const_path_ref, :const_path_field
          left = const_name(value[1])
          right = const_name(value[2])
          return "#{left}::#{right}" if left && right
        when :aref
          return const_name(value[1])
        end
      end
      nil
    end

    def qualify(namespace, name)
      name = name.to_s.delete_prefix("::")
      return name if namespace.empty? || name.include?("::")

      "#{namespace}::#{name}"
    end

    def parameter_names(params)
      tokens = []
      collect_tokens(params).each do |token|
        tokens << token[1] if %i[@ident @const].include?(token[0])
      end
      tokens.uniq
    end

    def collect_tokens(value, result = [])
      if token?(value)
        result << value
      elsif value.is_a?(Array)
        value.each { |child| collect_tokens(child, result) }
      end
      result
    end

    def first_argument(arguments)
      return nil unless arguments.is_a?(Array)
      return arguments[1].is_a?(Array) ? arguments[1].first : arguments[1] if arguments[0] == :args_add_block
      return first_argument(arguments[1]) if arguments[0] == :arg_paren

      arguments
    end

    def literal_string(value)
      return nil unless value.is_a?(Array)
      return value[1] if value[0] == :@tstring_content

      if value[0] == :string_literal
        content = value[1]
        return nil unless content.is_a?(Array) && content[0] == :string_content

        parts = content[1..].to_a
        return nil if parts.any? { |part| part.is_a?(Array) && part[0] == :string_embexpr }

        text = parts.filter_map { |part| part[1] if part.is_a?(Array) && part[0] == :@tstring_content }.join
        return text unless text.empty? && parts.any?
      end
      nil
    end

    def call_name_token(call)
      return nil unless call.is_a?(Array)

      case call[0]
      when :fcall, :vcall
        call[1]
      when :call
        call[3]
      when :command_call
        call[3]
      when :command
        call[1]
      end
    end

    def expression_for(value)
      tokens = collect_tokens(value)
      return "" if tokens.empty?

      tokens.map { |token| token[1].to_s }.join
    end

    def source_expression(value)
      tokens = collect_tokens(value)
      return "" if tokens.empty?

      first = tokens.first
      last = tokens.last
      start_line, start_column = first[2]
      end_line, end_column = last[2]
      end_column += last[1].to_s.bytesize
      lines = @source_lines[(start_line - 1)..(end_line - 1)].to_a
      return expression_for(value) if lines.empty?

      if lines.length == 1
        return lines.first.to_s.byteslice(start_column...end_column).to_s.strip
      end

      lines[0] = lines[0].to_s.byteslice(start_column..).to_s
      lines[-1] = lines[-1].to_s.byteslice(0...end_column).to_s
      lines.join.strip
    end

    def span_for(token)
      return nil unless token?(token)

      line, column = token[2]
      text = token[1].to_s
      {
        "start_line" => line,
        "start_column" => column + 1,
        "end_line" => line,
        "end_column" => column + text.length + 1,
        "path" => @current_path
      }
    end
  end
end
