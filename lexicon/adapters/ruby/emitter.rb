# frozen_string_literal: true

require "fileutils"

module LexiconRuby
  module JsonlEmitter
    private

    def write_facts
      incremental = !@changed_files.nil? || !@removed_files.nil?
      header = {
        "record" => "lexicon",
        "schema_version" => 1,
        "adapter_version" => LexiconRuby::Contract::VERSION,
        "language" => LexiconRuby::Contract::LANGUAGE,
        "mode" => incremental ? "incremental" : "full",
        "repository" => @repository_name
      }
      if incremental
        header["changed_files"] = (@changed_files || []).sort
        header["removed_files"] = (@removed_files || []).sort
        header["shared_complete"] = true
      end
      nodes = @nodes.values.sort_by do |record|
        [record["id"], record["kind"], record["path"], record["qualified_name"]]
      end
      owners = nodes.to_h { |record| [record["id"], direct_owner(record)] }
      records = nodes
      records += @edges.values.sort_by do |record|
        [record["source"], record["target"], record["relation"], *span_key(record)]
      end
      records += @unresolved.sort_by do |record|
        [record["source"], record["relation"], record["expression"], record["reason"], *span_key(record)]
      end
      if incremental
        selected = (@changed_files || []).to_h { |path| [path, true] }
        records = records.select { |record| include_record?(record, owners, selected) }
      end

      output_directory = File.dirname(@output)
      FileUtils.mkdir_p(output_directory) unless output_directory == "."
      File.open(@output, "wb") do |file|
        ([header] + records).each { |record| file.write("#{canonical_json(record)}\n") }
      end
      @output
    end

    def include_record?(record, owners, selected)
      owner = direct_owner(record)
      owner = owners[record["source"]] if owner.empty? && record["source"]
      owner.empty? || selected[owner]
    end

    def direct_owner(record)
      owner = record["owner"]
      return normalize_emission_path(owner) if owner && !owner.empty?

      span = record["span"]
      return normalize_emission_path(span["path"]) if span && span["path"]
      return normalize_emission_path(record["path"]) if record["record"] == "node" && record["kind"] == "file"

      ""
    end
  end
end
