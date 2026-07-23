package main

import "runtime"

type ScanOptions struct {
	SemanticWorkers int
	SemanticShards  int
	MergeFanIn      int
}

func (options ScanOptions) normalized(fileCount int) ScanOptions {
	if options.SemanticWorkers < 1 {
		options.SemanticWorkers = 1
	}
	if options.SemanticShards < 1 {
		options.SemanticShards = options.SemanticWorkers
	}
	if fileCount > 0 && options.SemanticShards > fileCount {
		options.SemanticShards = fileCount
	}
	if options.SemanticShards < 1 {
		options.SemanticShards = 1
	}
	if options.SemanticWorkers > options.SemanticShards {
		options.SemanticWorkers = options.SemanticShards
	}
	if maximum := runtime.GOMAXPROCS(0); options.SemanticWorkers > maximum {
		options.SemanticWorkers = maximum
	}
	if options.MergeFanIn < 2 {
		options.MergeFanIn = 2
	}
	return options
}
