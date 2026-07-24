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
        opts.on("--changed-file PATH", "repository-relative file to emit") { |value| (options[:changed_files] ||= []) << value }
        opts.on("--removed-file PATH", "repository-relative removed file") { |value| (options[:removed_files] ||= []) << value }
      end

      parser.parse!(argv)
      missing = %i[repo output].reject { |key| options[key] }
      abort("missing required option(s): #{missing.map { |key| "--#{key}" }.join(", ")}\n#{parser}") unless missing.empty?
      LexiconRubyAdapter.new(options[:repo], options[:output], options[:changed_files], options[:removed_files]).run
    rescue OptionParser::ParseError, ArgumentError => error
      abort(error.message)
    end
  end
end
