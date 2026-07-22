# frozen_string_literal: true

module LexiconRuby
  module CallHierarchy
    private

    def resolve_hierarchy
      @pending_extends.each do |source_id, owner, reference, span|
        info = type_info(owner)
        target_name = resolve_constant_name(reference, info&.declaration_namespace.to_s)
        target_name ||= external_type(reference.to_s.delete_prefix("::"), kind: "type")
        info.superclass_ref = target_name if info
        target_id = type_info(target_name)&.ids&.first
        add_edge(source_id, target_id, "extends", span) if target_id
      end

      @pending_mixin_edges.each do |owner, source_id, mode, references, span, namespace|
        info = type_info(owner)
        references.each do |reference|
          target_name = resolve_constant_name(reference, namespace)
          target_name ||= external_type(reference.to_s.delete_prefix("::"), kind: "module")
          collection = case mode
                       when "include" then info&.includes
                       when "prepend" then info&.prepends
                       else info&.extends_modules
                       end
          if collection
            collection.delete(reference)
            collection << target_name unless collection.include?(target_name)
          end
          target_id = type_info(target_name)&.ids&.first
          add_edge(source_id, target_id, "includes", span, { "mode" => mode }) if target_id
        end
      end
      rebuild_mixin_hosts
    end

    def external_type(name, kind:)
      clean = name.empty? ? "<anonymous>" : name
      return clean if @types.key?(clean)

      node_id = add_node(
        kind: kind,
        name: clean.split("::").last,
        path: "<external>",
        qualified_name: clean,
        canonical: "external\0#{kind}\0#{clean}",
        attributes: { "external" => true }
      )
      register_type(clean, kind, node_id, "", external: true)
      clean
    end

    def rebuild_mixin_hosts
      @mixin_hosts.clear
      @types.each_value do |info|
        next if info.external

        (info.includes + info.prepends).each { |mixin| @mixin_hosts[mixin] << info.name }
      end
    end

    def lookup_bare(owner, singleton, name, _source_id)
      return @method_definitions.dig("", false, name).to_a if owner.to_s.empty?

      candidates = singleton ? lookup_singleton(owner, name) : lookup_instance(owner, name)
      if candidates.empty? && type_info(owner)&.kind == "module"
        @mixin_hosts[owner].each { |host| candidates.concat(lookup_instance(host, name)) }
      end
      candidates.uniq
    end

    def lookup_instance(owner, name)
      lookup_chain(instance_chain(owner), false, name)
    end

    def lookup_singleton(owner, name)
      direct = @method_definitions.dig(owner, true, name).to_a
      return direct unless direct.empty?

      info = type_info(owner)
      return [] unless info

      info.extends_modules.reverse_each do |mixin|
        candidates = lookup_chain(instance_chain(mixin), false, name)
        return candidates unless candidates.empty?
      end
      superclass = info.superclass_ref
      superclass ? lookup_singleton(superclass, name) : []
    end

    def lookup_chain(chain, singleton, name)
      chain.each do |type_name|
        next if @undefined_methods[[type_name, singleton]].include?(name)

        candidates = @method_definitions.dig(type_name, singleton, name).to_a
        alias_name = @aliases[[type_name, singleton, name]]
        candidates = @method_definitions.dig(type_name, singleton, alias_name).to_a if candidates.empty? && alias_name
        return candidates unless candidates.empty?
      end
      []
    end

    def instance_chain(owner, seen = Set.new)
      return [] if owner.nil? || seen.include?(owner)

      next_seen = seen | [owner]
      info = type_info(owner)
      return [owner] unless info

      chain = []
      info.prepends.reverse_each { |mixin| chain.concat(instance_chain(mixin, next_seen)) }
      chain << owner
      info.includes.reverse_each { |mixin| chain.concat(instance_chain(mixin, next_seen)) }
      chain.concat(instance_chain(info.superclass_ref, next_seen)) if info.superclass_ref
      chain.uniq
    end

    def resolve_super_targets(call)
      method = @methods[call.source]
      return Set.new unless method&.owner

      if method.singleton
        superclass = type_info(method.owner)&.superclass_ref
        return Set.new(superclass ? lookup_singleton(superclass, method.name) : [])
      end

      chains = if type_info(method.owner)&.kind == "module" && !@mixin_hosts[method.owner].empty?
                 @mixin_hosts[method.owner].map { |host| instance_chain(host) }
               else
                 [instance_chain(method.owner)]
               end
      targets = Set.new
      chains.each do |chain|
        index = chain.index(method.owner)
        next unless index

        targets.merge(lookup_chain(chain[(index + 1)..].to_a, false, method.name))
      end
      targets
    end

    def external_dispatch?(owner, seen = Set.new)
      return false unless owner
      return false if seen.include?(owner)

      info = type_info(owner)
      return false unless info
      return true if info.external

      next_seen = seen | [owner]
      external_dispatch?(info.superclass_ref, next_seen) ||
        (info.includes + info.prepends).any? { |mixin| external_dispatch?(mixin, next_seen) }
    end
  end
end
