# frozen_string_literal: true

module LexiconRuby
  module RipperSyntax
    private

    def const_name(value)
      return nil unless value
      return value[1] if token?(value) && value[0] == :@const

      return nil unless value.is_a?(Array)

      case value[0]
      when :const_ref, :var_ref, :var_field
        const_name(value[1])
      when :top_const_ref
        name = const_name(value[1])
        name && "::#{name}"
      when :const_path_ref, :const_path_field
        left = const_name(value[1])
        right = const_name(value[2])
        left && right ? "#{left}::#{right}" : nil
      when :aref
        const_name(value[1])
      end
    end

    def self_reference?(value)
      token = first_token(value)
      token && token[0] == :@kw && token[1] == "self"
    end

    def qualify(namespace, name)
      raw = name.to_s
      return raw.delete_prefix("::") if raw.start_with?("::")
      return raw if namespace.empty? || raw.include?("::")

      "#{namespace}::#{raw}"
    end

    def lexical_constant_candidates(namespace, name)
      raw = name.to_s
      return [] if raw.empty?
      return [raw.delete_prefix("::")] if raw.start_with?("::")

      prefixes = namespace.to_s.split("::")
      candidates = []
      prefixes.length.downto(0) do |length|
        prefix = prefixes.first(length).join("::")
        candidates << (prefix.empty? ? raw : "#{prefix}::#{raw}")
      end
      candidates.uniq
    end

    def parameter_spec(params)
      node = params
      node = node[1] if sexp_node?(node, :paren)
      return { positional: [], optional: [], keywords: {}, rest: nil, keyword_rest: nil, block: nil } unless sexp_node?(node, :params)

      required, optional, rest, post, keywords, keyword_rest, block = node[1..7]
      positional = Array(required).filter_map { |token| token_value(token) }
      positional.concat(Array(post).filter_map { |token| token_value(token) })
      optional_values = Array(optional).filter_map do |entry|
        next unless entry.is_a?(Array)

        name = token_value(entry[0])
        name && [name, entry[1]]
      end
      keyword_values = {}
      Array(keywords).each do |entry|
        next unless entry.is_a?(Array)

        token = entry[0]
        name = token_value(token).to_s.delete_suffix(":")
        keyword_values[name] = entry[1] unless name.empty?
      end
      {
        positional: positional,
        optional: optional_values,
        keywords: keyword_values,
        rest: parameter_token_name(rest),
        keyword_rest: parameter_token_name(keyword_rest),
        block: parameter_token_name(block)
      }
    end

    def parameter_token_name(value)
      return nil unless value.is_a?(Array)

      token_value(value[1])
    end

    def parameter_names(params)
      spec = parameter_spec(params)
      names = spec[:positional] + spec[:optional].map(&:first) + spec[:keywords].keys
      names << spec[:rest] if spec[:rest]
      names << spec[:keyword_rest] if spec[:keyword_rest]
      names << spec[:block] if spec[:block]
      names
    end

    def argument_parts(arguments)
      node = arguments
      node = node[1] if sexp_node?(node, :arg_paren)
      return [[], {}] if node.nil? || node == []

      values = if sexp_node?(node, :args_add_block)
                 Array(node[1])
               elsif sexp_node?(node, :args_add_star)
                 Array(node[1]) + Array(node[2])
               elsif node.is_a?(Array) && node.first.is_a?(Array)
                 node
               else
                 [node]
               end
      positional = []
      keywords = {}
      values.each do |value|
        if sexp_node?(value, :bare_assoc_hash) || sexp_node?(value, :hash)
          association_list(value).each do |association|
            key, item = association[1], association[2]
            label = token_value(key)
            if label&.end_with?(":")
              keywords[label.delete_suffix(":")] = item
            else
              positional << value
              break
            end
          end
        else
          positional << value
        end
      end
      [positional, keywords]
    end

    def association_list(value)
      return [] unless value.is_a?(Array)

      if sexp_node?(value, :bare_assoc_hash)
        Array(value[1])
      elsif sexp_node?(value, :hash) && sexp_node?(value[1], :assoclist_from_args)
        Array(value[1][1])
      else
        []
      end
    end

    def symbol_names(arguments)
      positional, = argument_parts(arguments)
      positional.filter_map { |value| literal_symbol(value) }
    end

    def literal_symbol(value)
      return nil unless value.is_a?(Array)

      if sexp_node?(value, :symbol_literal)
        token_value(first_token(value[1]))
      elsif token?(value) && value[0] == :@label
        value[1].delete_suffix(":")
      end
    end

    def variable_target(value)
      return nil unless value.is_a?(Array)

      node = value
      node = node[1] if %i[var_field const_path_field].include?(node[0]) && node[0] == :var_field
      token = first_token(node)
      return nil unless token

      kind = case token[0]
             when :@ident then :local
             when :@ivar then :instance
             when :@cvar then :class_variable
             when :@gvar then :global
             when :@const then :constant
             end
      return nil unless kind

      { kind: kind, name: token[1], token: token, constant: const_name(value) }
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
      positional, = argument_parts(arguments)
      positional.first
    end

    def literal_string(value)
      return nil unless value.is_a?(Array)
      return value[1] if value[0] == :@tstring_content

      return nil unless value[0] == :string_literal

      content = value[1]
      return nil unless content.is_a?(Array) && content[0] == :string_content

      parts = content[1..].to_a
      return nil if parts.any? { |part| sexp_node?(part, :string_embexpr) }

      text = parts.filter_map { |part| part[1] if sexp_node?(part, :@tstring_content) }.join
      text unless text.empty? && parts.any?
    end

    def expression_for(value)
      tokens = collect_tokens(value)
      return "" if tokens.empty?

      tokens.map { |token| token[1].to_s }.join
    end

    def source_expression(value)
      tokens = collect_tokens(value)
      return "" if tokens.empty?

      first = tokens.min_by { |token| token_position(token) }
      last = tokens.max_by { |token| token_position(token) }
      start_line, start_column = first[2]
      end_line, end_column = last[2]
      end_column += last[1].to_s.bytesize
      lines = @source_lines[(start_line - 1)..(end_line - 1)].to_a
      return expression_for(value) if lines.empty?

      return lines.first.to_s.byteslice(start_column...end_column).to_s.strip if lines.length == 1

      lines[0] = lines[0].to_s.byteslice(start_column..).to_s
      lines[-1] = lines[-1].to_s.byteslice(0...end_column).to_s
      lines.join.strip
    end

    def span_for_expression(value)
      tokens = collect_tokens(value)
      return nil if tokens.empty?

      first = tokens.min_by { |token| token_position(token) }
      last = tokens.max_by { |token| token_position(token) }
      start_line, start_column = first[2]
      end_line, end_column = last[2]
      {
        "start_line" => start_line,
        "start_column" => start_column + 1,
        "end_line" => end_line,
        "end_column" => end_column + last[1].to_s.length + 1,
        "path" => @current_path
      }
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

    def position_for(value)
      token_position(first_token(value))
    end
  end
end
