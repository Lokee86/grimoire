# frozen_string_literal: true

require "json"

module LexiconRuby
  module DependencySemantics
    private

    def dependency_attributes(category, source, constraint = "", path: false)
      {
        "build" => category == "build",
        "category" => category,
        "constraint" => constraint,
        "dev" => ["development", "test"].include?(category),
        "optional" => category == "optional",
        "path" => path,
        "peer" => category == "peer",
        "source" => source
      }
    end

    def dependency_target(name, local_path = "")
      normalized = (local_path.empty? ? name : local_path).tr("\\", "/")
      identity = "dependency:ruby:#{normalized}"
      add_node(
        kind: "module",
        name: normalized,
        path: local_path.empty? ? ".lexicon/dependencies/ruby/#{normalized}" : normalized,
        qualified_name: identity,
        canonical: identity,
        attributes: { "dependency" => true, "ecosystem" => "ruby" }
      )
    end

    def safe_dependency_path(value)
      normalized = value.to_s.tr("\\", "/")
      return "" if normalized.empty? || normalized.start_with?("/") || normalized == ".." || normalized.start_with?("../")
      cleaned = File.expand_path(normalized, "/")
      return "" if cleaned == "/.." || cleaned.start_with?("/../")

      cleaned.delete_prefix("/")
    end

    def add_manifest_dependency(name, constraint, source, category, local_path: "")
      repository_id = @nodes.values.find { |record| record["kind"] == "repository" }["id"]
      add_edge(repository_id, dependency_target(name, local_path), "depends-on", nil,
               dependency_attributes(category, source, constraint, path: !local_path.empty?))
    end

    def add_dependency_facts
      parse_gemfile(File.join(@repo, "Gemfile")) if File.file?(File.join(@repo, "Gemfile"))
      Dir.glob(File.join(@repo, "**", "*.gemspec")).sort.each do |filename|
        next if excluded_path?(normalize_path(relative_path_for(filename)))
        parse_gemspec(filename)
      end
      resolve_local_dependencies
    end

    def parse_gemfile(filename)
      category = "runtime"
      File.readlines(filename, encoding: "UTF-8").each do |line|
        text = line.split("#", 2).first.strip
        if text =~ /\bgroup\s+[^\{]+(?:\{|do)/ && text =~ /:(development|test|build|optional|peer)/
          category = Regexp.last_match(1)
          next
        end
        category = "runtime" if text == "end"
        match = text.match(/^gem\s+(["'])([^"']+)\1(.*)$/)
        next unless match
        name = match[2]
        tail = match[3]
        constraint = tail.scan(/["']([^"']+)["']/).flatten.first.to_s
        local = tail.match(/(?:path|git):\s*["']([^"']+)["']/)
        local_path = local && local[0].include?("path:") ? safe_dependency_path(local[1]) : ""
        add_manifest_dependency(name, constraint, "Gemfile:gem", category, local_path: local_path)
      end
    rescue ArgumentError, IOError
      nil
    end

    def parse_gemspec(filename)
      File.readlines(filename, encoding: "UTF-8").each do |line|
        text = line.split("#", 2).first.strip
        match = text.match(/\badd_(development_|runtime_)?dependency\s+(["'])([^"']+)\2(.*)$/)
        next unless match
        category = match[1] == "development_" ? "development" : "runtime"
        constraint = match[4].scan(/["']([^"']+)["']/).flatten.first.to_s
        source = "#{normalize_path(relative_path_for(filename))}:#{match[1] || "runtime_"}dependency"
        add_manifest_dependency(match[3], constraint, source, category)
      end
    rescue ArgumentError, IOError
      nil
    end

    def record_local_dependency(path, required)
      @pending_local_dependencies << [path, required]
    end

    def resolve_local_dependencies
      requests = @pending_local_dependencies.dup
      @nodes.values.each do |record|
        attributes = record["attributes"] || {}
        requests << [record["path"], attributes["target"]] if record["kind"] == "import" && attributes["loader"] == "require_relative"
      end
      requests.uniq.each do |source_path, required|
        next if required.nil? || required.empty? || required.start_with?("/")
        candidate = normalize_path(File.join(File.dirname(source_path), required))
        candidates = [candidate, "#{candidate}.rb", File.join(candidate, "init.rb")].map { |value| normalize_path(value) }
        target_path = candidates.find { |value| @files.key?(value) }
        source = @module_ids[source_path]
        target = target_path && @module_ids[target_path]
        next unless source && target
        add_edge(source, target, "depends-on", nil, dependency_attributes("local", required, path: true))
      end
    end
  end
end
