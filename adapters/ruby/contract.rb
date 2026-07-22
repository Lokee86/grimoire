# frozen_string_literal: true

require "digest"
require "json"

module LexiconRuby
  module Contract
    VERSION = "0.1.0"
    LANGUAGE = "ruby"
    EXCLUDED_DIRECTORIES = %w[
      .git .worktrees .workingtrees .warlock
      .bundle vendor node_modules target build dist tmp log coverage
    ].freeze
    METAPROGRAMMING_CALLS = %w[
      class_eval module_eval instance_eval define_method define_singleton_method
      const_set class_variable_set eval instance_exec send public_send
    ].freeze

    private

    def node_id(kind, canonical)
      "sha256:#{Digest::SHA256.hexdigest("lexicon:v1\0#{LANGUAGE}\0#{kind}\0#{canonical}")}"
    end

    def content_id(content)
      "sha256:#{Digest::SHA256.hexdigest(content)}"
    end

    def span_key(record)
      span = record["span"] || {}
      [span["path"] || "", span["start_line"] || 0, span["start_column"] || 0,
       span["end_line"] || 0, span["end_column"] || 0]
    end

    def canonical_record(record)
      canonical_json(record)
    end

    def canonical_json(value)
      JSON.generate(sort_hashes(value))
    end

    def sort_hashes(value)
      case value
      when Hash
        value.keys.sort.each_with_object({}) { |key, result| result[key] = sort_hashes(value[key]) }
      when Array
        value.map { |item| sort_hashes(item) }
      else
        value
      end
    end

    def sexp_node?(value, tag)
      value.is_a?(Array) && value.first == tag
    end

    def token?(value)
      value.is_a?(Array) && value.length == 3 && value.first.is_a?(Symbol) && value.first.to_s.start_with?("@")
    end

    def first_token(value)
      return value if token?(value)
      return nil unless value.is_a?(Array)

      value.each do |child|
        token = first_token(child)
        return token if token
      end
      nil
    end

    def token_value(token)
      token?(token) ? token[1] : nil
    end

    def token_position(token)
      token?(token) ? token[2] : [0, 0]
    end
  end
end
