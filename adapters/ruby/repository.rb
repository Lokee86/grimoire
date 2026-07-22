# frozen_string_literal: true

require "ripper"

module LexiconRuby
  module RepositoryDiscovery
    def run
      raise ArgumentError, "repository does not exist: #{@repo}" unless Dir.exist?(@repo)

      repository_id = add_node(
        kind: "repository",
        name: @repository_name,
        path: ".",
        qualified_name: @repository_name,
        canonical: @repository_name
      )
      root_directory_id = add_directory(".")
      add_edge(repository_id, root_directory_id, "contains")

      ruby_files.each { |relative_path| scan_file(relative_path) }
      resolve_superclasses
      write_facts
    end

    private

    def ruby_files
      Dir.glob(File.join(@repo, "**", "*.rb"), File::FNM_EXTGLOB).filter_map do |absolute_path|
        next unless File.file?(absolute_path)

        relative_path = normalize_path(relative_path_for(absolute_path))
        next if excluded_path?(relative_path)

        relative_path
      end.sort
    end

    def relative_path_for(path)
      path.delete_prefix("#{@repo}#{File::SEPARATOR}").delete_prefix("#{@repo}/")
    end

    def normalize_path(path)
      path.tr("\\", "/")
    end

    def excluded_path?(relative_path)
      relative_path.split("/").any? { |component| LexiconRuby::Contract::EXCLUDED_DIRECTORIES.include?(component) }
    end

    def scan_file(relative_path)
      absolute_path = File.join(@repo, relative_path)
      source = File.binread(absolute_path).force_encoding(Encoding::UTF_8)
      @source_lines = source.lines
      @current_path = relative_path

      file_id = add_node(
        kind: "file",
        name: File.basename(relative_path),
        path: relative_path,
        qualified_name: relative_path,
        canonical: relative_path,
        content_id: content_id(source)
      )
      @files[relative_path] = file_id
      parent_directory_id = add_directory(File.dirname(relative_path))
      add_edge(parent_directory_id, file_id, "contains")

      sexp = Ripper.sexp(source)
      if sexp.nil?
        add_unresolved(
          source: file_id,
          relation: "parses",
          expression: relative_path,
          reason: "unsupported-form"
        )
        return
      end

      visit(
        sexp,
        {
          parent_id: file_id,
          source_id: file_id,
          namespace: "",
          owner: nil,
          singleton: false,
          method_id: nil,
          scope_id: file_id,
          branch_depth: 0,
          mixin_hook: nil
        }
      )
    rescue ArgumentError => error
      add_unresolved(
        source: @files[relative_path] || add_node(
          kind: "file",
          name: File.basename(relative_path),
          path: relative_path,
          qualified_name: relative_path,
          canonical: relative_path
        ),
        relation: "parses",
        expression: relative_path,
        reason: "unsupported-form",
        attributes: { "message" => error.message }
      )
    end

    def add_directory(relative_path)
      relative_path = normalize_path(relative_path)
      return @directories[relative_path] if @directories.key?(relative_path)

      directory_id = add_node(
        kind: "directory",
        name: relative_path == "." ? @repository_name : File.basename(relative_path),
        path: relative_path,
        qualified_name: relative_path,
        canonical: relative_path
      )
      @directories[relative_path] = directory_id

      unless relative_path == "."
        parent = File.dirname(relative_path)
        parent = "." if parent == ""
        parent_id = add_directory(parent)
        add_edge(parent_id, directory_id, "contains")
      end

      directory_id
    end
  end
end
