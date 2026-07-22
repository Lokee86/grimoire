#!/usr/bin/env ruby
# frozen_string_literal: true

require_relative "contract"
require_relative "relationships"
require_relative "ripper_syntax"
require_relative "ripper_extractor"
require_relative "repository"
require_relative "emitter"
require_relative "model"
require_relative "cli"

LexiconRuby::CLI.run(ARGV) if __FILE__ == $PROGRAM_NAME
