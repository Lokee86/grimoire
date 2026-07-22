# frozen_string_literal: true

require "json"
require "minitest/autorun"
require "open3"
require "rbconfig"
require "tmpdir"

class RubyAdapterTest < Minitest::Test
  ROOT = File.expand_path("..", __dir__)
  ADAPTER = File.join(ROOT, "lexicon_ruby.rb")
  FIXTURE = File.join(__dir__, "fixtures", "sample")
  ID_PATTERN = /\Asha256:[0-9a-f]{64}\z/

  def run_adapter(output)
    stdout, stderr, status = Open3.capture3(
      RbConfig.ruby,
      ADAPTER,
      "--repo", FIXTURE,
      "--output", output
    )
    assert status.success?, "adapter failed: #{stderr}\n#{stdout}"
  end

  def records_for(output)
    File.readlines(output, chomp: true, encoding: "UTF-8").map { |line| JSON.parse(line) }
  end

  def test_extracts_nested_declarations_methods_requires_and_inheritance
    Dir.mktmpdir do |directory|
      output = File.join(directory, "facts.jsonl")
      run_adapter(output)
      records = records_for(output)
      nodes = records.select { |record| record["record"] == "node" }
      edges = records.select { |record| record["record"] == "edge" }
      unresolved = records.select { |record| record["record"] == "unresolved" }

      assert_equal "lexicon", records.first["record"]
      assert_equal "ruby", records.first["language"]
      assert_equal "sample", records.first["repository"]
      assert nodes.any? { |record| record["kind"] == "repository" }
      assert nodes.any? { |record| record["kind"] == "directory" && record["path"] == "lib" }
      assert nodes.any? { |record| record["kind"] == "file" && record["path"] == "lib/sample.rb" }
      assert nodes.any? { |record| record["kind"] == "module" && record["qualified_name"] == "Outer::Inner" }
      assert nodes.any? { |record| record["kind"] == "type" && record["qualified_name"] == "Outer::Inner::Child" }
      assert nodes.any? { |record| record["kind"] == "method" && record["qualified_name"] == "Outer::Inner::Child#run" }
      assert nodes.any? { |record| record["kind"] == "constant" && record["qualified_name"] == "Outer::VERSION" }
      assert nodes.any? { |record| record["kind"] == "import" && record["attributes"]["target"] == "json" }
      assert edges.any? { |record| record["relation"] == "contains" }
      assert edges.any? { |record| record["relation"] == "defines" }
      assert edges.any? { |record| record["relation"] == "imports" }
      assert edges.any? { |record| record["relation"] == "extends" }
      assert unresolved.any? { |record| record["relation"] == "imports" && record["reason"] == "dynamic-target" }
      refute nodes.any? { |record| record["path"].include?("vendor") }
      refute nodes.any? { |record| record["path"].include?((".git")) }
    end
  end

  def test_ids_are_unique_and_records_are_deterministically_sorted
    Dir.mktmpdir do |directory|
      first = File.join(directory, "first.jsonl")
      second = File.join(directory, "second.jsonl")
      run_adapter(first)
      run_adapter(second)
      assert_equal File.binread(first), File.binread(second)

      records = records_for(first)
      facts = records.drop(1)
      node_ids = facts.select { |record| record["record"] == "node" }.map { |record| record["id"] }
      assert_equal node_ids.uniq, node_ids
      node_ids.each { |id| assert_match ID_PATTERN, id }
      facts.each { |record| assert_equal record.keys.sort, record.keys }

      node_records = facts.select { |record| record["record"] == "node" }
      edge_records = facts.select { |record| record["record"] == "edge" }
      unresolved_records = facts.select { |record| record["record"] == "unresolved" }
      assert_equal node_records.sort_by { |record| [record["id"], record["kind"], record["path"], record["qualified_name"]] }, node_records
      assert_equal edge_records.sort_by { |record| [record["source"], record["target"], record["relation"], *span_key(record)] }, edge_records
      assert_equal unresolved_records.sort_by { |record| [record["source"], record["relation"], record["expression"], record["reason"], *span_key(record)] }, unresolved_records
    end
  end

  private

  def span_key(record)
    span = record["span"] || {}
    [span["path"] || "", span["start_line"] || 0, span["start_column"] || 0,
     span["end_line"] || 0, span["end_column"] || 0]
  end
end
