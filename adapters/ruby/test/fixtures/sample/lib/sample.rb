require "json"
require_relative "support"
require(dynamic_target)

module Outer
  VERSION = "1.0"

  module Inner
    class Base
      def base(value)
        value
      end
    end

    class Child < Base
      def initialize(value)
        @value = value
      end

      def run
        @value
      end

      def self.build(value)
        new(value)
      end
    end

    class Versioned < Base[1]
    end
  end
end
