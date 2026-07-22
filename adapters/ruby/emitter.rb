# frozen_string_literal: true

require "fileutils"

module LexiconRuby
  module JsonlEmitter
    private

    def write_facts
      header = {
        "record" => "lexicon",
        "schema_version" => 1,
        "adapter_version" => LexiconRuby::Contract::VERSION,
        "language" => LexiconRuby::Contract::LANGUAGE,
        "repository" => @repository_name
      }
      records = @nodes.values.sort_by do |record|
        [record["id"], record["kind"], record["path"], record["qualified_name"]]
      end
      records += @edges.values.sort_by do |record|
        [record["source"], record["target"], record["relation"], *span_key(record)]
      end
      records += @unresolved.sort_by do |record|
        [record["source"], record["relation"], record["expression"], record["reason"], *span_key(record)]
      end

      output_directory = File.dirname(@output)
      FileUtils.mkdir_p(output_directory) unless output_directory == "."
      File.open(@output, "wb") do |file|
        ([header] + records).each { |record| file.write("#{canonical_json(record)}\n") }
      end
      @output
    end
  end
end
