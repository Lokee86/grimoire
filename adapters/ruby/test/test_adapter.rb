# frozen_string_literal: true

require "fileutils"
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

  def run_adapter(output, repo: FIXTURE)
    stdout, stderr, status = Open3.capture3(
      RbConfig.ruby,
      ADAPTER,
      "--repo", repo,
      "--output", output
    )
    assert status.success?, "adapter failed: #{stderr}\n#{stdout}"
  end

  def records_for(output)
    File.readlines(output, chomp: true, encoding: "UTF-8").map { |line| JSON.parse(line) }
  end

  def test_extracts_declarations_imports_and_inheritance
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
      %w[contains defines imports extends].each do |relation|
        assert edges.any? { |record| record["relation"] == relation }
      end
      assert unresolved.any? { |record| record["relation"] == "imports" && record["reason"] == "dynamic-target" }
      refute unresolved.any? { |record| record["relation"] == "extends" }
      edge_keys = edges.map do |record|
        [record["source"], record["target"], record["relation"], span_key(record), record["attributes"]]
      end
      assert_equal edge_keys.uniq.length, edges.length
      refute nodes.any? { |record| record["path"].include?("vendor") }
      refute nodes.any? { |record| record["path"].include?(".git") }
    end
  end

  def test_ids_and_output_are_deterministic
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

      nodes = facts.select { |record| record["record"] == "node" }
      edges = facts.select { |record| record["record"] == "edge" }
      unresolved = facts.select { |record| record["record"] == "unresolved" }
      assert_equal nodes.sort_by { |record| [record["id"], record["kind"], record["path"], record["qualified_name"]] }, nodes
      assert_equal edges.sort_by { |record| [record["source"], record["target"], record["relation"], *span_key(record)] }, edges
      assert_equal unresolved.sort_by { |record| [record["source"], record["relation"], record["expression"], record["reason"], *span_key(record)] }, unresolved
    end
  end

  def test_preserves_callsites_blocks_and_ambiguity
    source = <<~RUBY
      class Local
        register :external_dsl

        def run
          helper
          unique(1)
          missing
          explicit.helper
          send(:helper)
          records.each do |record|
            helper
          end
          helper do
            helper
          end
        end

        def helper; end
        def unique(value); value; end
        def ambiguous; end
        def ambiguous; end
        def use_ambiguous; ambiguous; end
      end
    RUBY
    with_repository("local.rb" => source) do |records|
      nodes = node_index(records)
      run_id = id_for(nodes, "Local#run")
      direct_targets = call_targets(records, nodes, run_id, "calls")
      assert_equal 2, direct_targets.count("Local#helper")
      assert_equal 1, direct_targets.count("Local#unique")

      block_ids = nodes.values.select { |node| node["kind"] == "function" && node.dig("attributes", "block") }.map { |node| node["id"] }
      assert(records.any? do |record|
        record["record"] == "edge" && block_ids.include?(record["source"]) &&
          nodes[record["target"]]&.dig("qualified_name") == "Local#helper"
      end)

      unresolved = records.select { |record| record["record"] == "unresolved" }
      assert unresolved.any? { |record| record["expression"] == "missing" && record["reason"] == "missing-target" }
      assert unresolved.any? { |record| record["expression"] == "explicit.helper" && record["reason"] == "dynamic-target" }
      assert unresolved.any? do |record|
        record["relation"] == "defines" && record["expression"].start_with?("send")
      end

      redefined_id = id_for(nodes, "Local#use_ambiguous")
      assert_equal 1, call_targets(records, nodes, redefined_id, "calls").count("Local#ambiguous")
      refute(unresolved.any? do |record|
        record["source"] == redefined_id && record["reason"] == "ambiguous-target"
      end)
    end
  end

  def test_resolves_ruby_static_call_semantics
    source = <<~RUBY
      module Support
        def support_call
          base_helper
        end
      end

      module Installed
        def installed_call; end
      end

      class Framework; end
      Framework.include(Installed)

      class Base
        include Support
        def base_helper; end
        def inherited; end
      end

      class Product
        def initialize; end
        def work; end
      end

      module Factory
        module_function
        def build
          Product.new
        end
      end

      class SingletonFactory
        class << self
          def build
            Product.new
          end
        end
      end

      Result = Struct.new(:value) do
        def success?
          value
        end
      end

      class Child < Base
        def run
          support_call
          inherited
          Factory.build.work
          SingletonFactory.build.work
          result = Result.new(value: Product.new)
          result.success?
          around { base_helper }
        end

        def around
          yield
        end
      end

      class InstalledChild < Framework
        def run
          installed_call
        end
      end

      class Parent
        def execute; end
      end

      class SuperChild < Parent
        def execute
          super
        end
      end

      class Alpha
        def execute; end
      end

      class Beta
        def execute; end
      end

      def dispatch(value)
        value.execute
      end

      dispatch(Alpha.new)
      dispatch(Beta.new)
    RUBY
    with_repository("semantic.rb" => source) do |records|
      nodes = node_index(records)
      assert_call(records, nodes, "Child#run", "Support#support_call")
      assert_call(records, nodes, "Child#run", "Base#inherited")
      assert_call(records, nodes, "Support#support_call", "Base#base_helper")
      assert_call(records, nodes, "Child#run", "Factory.build")
      assert_call(records, nodes, "Child#run", "SingletonFactory.build")
      assert_operator call_targets(records, nodes, id_for(nodes, "Child#run"), "calls").count("Product#work"), :>=, 2
      assert_call(records, nodes, "Child#run", "Result#success?")
      assert_call(records, nodes, "Result#success?", "Result#value")
      assert_call(records, nodes, "InstalledChild#run", "Installed#installed_call")
      assert_call(records, nodes, "SuperChild#execute", "Parent#execute")

      dispatch_id = id_for(nodes, "dispatch")
      assert_equal ["Alpha#execute", "Beta#execute"], call_targets(records, nodes, dispatch_id, "possible-calls").sort
      refute(records.any? do |record|
        record["record"] == "unresolved" && record["reason"] == "missing-target" &&
          record.dig("span", "path") == "semantic.rb" && record["expression"] != "value.execute"
      end)
    end
  end

  private

  def with_repository(files)
    Dir.mktmpdir do |repository|
      files.each do |path, contents|
        absolute = File.join(repository, path)
        FileUtils.mkdir_p(File.dirname(absolute))
        File.write(absolute, contents)
      end
      output = File.join(repository, "facts.jsonl")
      run_adapter(output, repo: repository)
      yield records_for(output)
    end
  end

  def node_index(records)
    records.select { |record| record["record"] == "node" }.to_h { |record| [record["id"], record] }
  end

  def id_for(nodes, qualified_name)
    nodes.values.find { |node| node["qualified_name"] == qualified_name }&.fetch("id")
  end

  def call_targets(records, nodes, source_id, relation)
    records.filter_map do |record|
      next unless record["record"] == "edge" && record["source"] == source_id && record["relation"] == relation

      nodes[record["target"]]&.dig("qualified_name")
    end
  end

  def assert_call(records, nodes, source, target)
    assert_includes call_targets(records, nodes, id_for(nodes, source), "calls"), target
  end

  def span_key(record)
    span = record["span"] || {}
    [span["path"] || "", span["start_line"] || 0, span["start_column"] || 0,
     span["end_line"] || 0, span["end_column"] || 0]
  end
end
