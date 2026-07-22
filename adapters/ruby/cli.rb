# frozen_string_literal: true

require "optparse"

module LexiconRuby
  module CLI
    module_function

    def run(argv)
      options = {}
      parser = OptionParser.new do |opts|
        opts.banner = "Usage: lexicon_ruby.rb --repo PATH --output PATH"
        opts.on("--repo PATH", "repository root") { |value| options[:repo] = value }
        opts.on("--output PATH", "JSONL output path") { |value| options[:output] = value }
      end

      parser.parse!(argv)
      missing = %i[repo output].reject { |key| options[key] }
      abort("missing required option(s): #{missing.map { |key| "--#{key}" }.join(", ")}\n#{parser}") unless missing.empty?
      LexiconRubyAdapter.new(options[:repo], options[:output]).run
    rescue OptionParser::ParseError, ArgumentError => error
      abort(error.message)
    end
  end
end
