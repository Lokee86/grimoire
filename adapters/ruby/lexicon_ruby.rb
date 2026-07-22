#!/usr/bin/env ruby
# frozen_string_literal: true

require_relative "contract"
require_relative "semantic_model"
require_relative "relationships"
require_relative "ripper_syntax"
require_relative "ripper_closures"
require_relative "ripper_calls"
require_relative "ripper_declarations"
require_relative "ripper_assignments"
require_relative "ripper_extractor"
require_relative "call_shapes"
require_relative "call_hierarchy"
require_relative "call_flow"
require_relative "call_emission"
require_relative "call_resolution"
require_relative "repository"
require_relative "emitter"
require_relative "model"
require_relative "cli"

LexiconRuby::CLI.run(ARGV) if __FILE__ == $PROGRAM_NAME
