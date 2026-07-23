# frozen_string_literal: true

require "set"

module LexiconRuby
  TypeInfo = Struct.new(
    :name,
    :kind,
    :ids,
    :declaration_namespace,
    :superclass_ref,
    :includes,
    :prepends,
    :extends_modules,
    :external,
    keyword_init: true
  )

  MethodInfo = Struct.new(
    :id,
    :owner,
    :name,
    :singleton,
    :parameters,
    :body,
    :namespace,
    :path,
    :position,
    :synthetic,
    keyword_init: true
  )

  CallSite = Struct.new(
    :key,
    :source,
    :owner,
    :singleton,
    :namespace,
    :receiver,
    :receiver_expression,
    :name,
    :arguments,
    :keywords,
    :block_id,
    :expression,
    :span,
    :node,
    :kind,
    keyword_init: true
  )

  AssignmentInfo = Struct.new(
    :scope,
    :owner,
    :singleton,
    :name,
    :kind,
    :value,
    :position,
    :branch_dependent,
    keyword_init: true
  )

  BlockInfo = Struct.new(
    :id,
    :owner,
    :singleton,
    :parameters,
    :call_key,
    :body,
    :path,
    keyword_init: true
  )

  class ValueShape
    attr_reader :instances, :singletons, :callables, :elements

    def initialize(instances: nil, singletons: nil, callables: nil, elements: nil)
      @instances = Set.new(instances)
      @singletons = Set.new(singletons)
      @callables = Set.new(callables)
      @elements = Set.new(elements)
    end

    def empty?
      @instances.empty? && @singletons.empty? && @callables.empty? && @elements.empty?
    end

    def merge(other)
      return self if other.nil? || other.empty?
      return other if empty?

      ValueShape.new(
        instances: @instances | other.instances,
        singletons: @singletons | other.singletons,
        callables: @callables | other.callables,
        elements: @elements | other.elements
      )
    end

    def element_shape
      ValueShape.new(instances: @elements)
    end

    def ==(other)
      other.is_a?(ValueShape) &&
        @instances == other.instances &&
        @singletons == other.singletons &&
        @callables == other.callables &&
        @elements == other.elements
    end

    alias eql? ==

    def hash
      [@instances, @singletons, @callables, @elements].hash
    end
  end

  EMPTY_SHAPE = ValueShape.new.freeze

  BUILTIN_CALLS = Set.new(%w[
    raise fail abort exit exit! throw catch loop sleep rand srand
    Integer Float String Array Hash Rational Complex
    puts print printf p pp warn sprintf format
    require require_relative load autoload autoload?
    block_given? caller caller_locations binding eval exec fork spawn system
    lambda proc at_exit select test trace_var untrace_var
    __dir__ __method__ __callee__ freeze respond_to? respond_to_missing?
    instance_variable_get instance_variable_set instance_variable_defined?
    remove_instance_variable method public_method singleton_class class
    private_class_method public_class_method [] []=
  ]).freeze

  EXTERNAL_BARE_CALLS = Set.new(%w[
    render redirect_to head params cookies session request response
    action_name controller_path
    before_action after_action around_action skip_before_action rescue_from
    validates validate belongs_to has_one has_many scope enum serialize
    attribute attribute_alias accepts_nested_attributes_for
    test setup teardown assert assert_equal assert_nil assert_not_nil
    assert_includes refute refute_equal refute_includes assert_response
    assert_predicate assert_empty assert_match assert_kind_of assert_raises
    get post put patch delete add_index add_reference add_column remove_column
    remove_index change_column_null create_table drop_table change_table
    reversible up down execute say say_with_time
  ]).freeze
end
