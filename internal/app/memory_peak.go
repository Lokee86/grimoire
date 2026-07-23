package app

import (
	"runtime"
	"sync/atomic"
	"time"
)

type memoryPeakSampler struct {
	peak atomic.Uint64
	stop chan struct{}
	done chan struct{}
}

func startMemoryPeakSampler() *memoryPeakSampler {
	sampler := &memoryPeakSampler{stop: make(chan struct{}), done: make(chan struct{})}
	sampler.sample()
	go func() {
		defer close(sampler.done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sampler.sample()
			case <-sampler.stop:
				sampler.sample()
				return
			}
		}
	}()
	return sampler
}

func (sampler *memoryPeakSampler) sample() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	for current := sampler.peak.Load(); stats.Sys > current; current = sampler.peak.Load() {
		if sampler.peak.CompareAndSwap(current, stats.Sys) {
			break
		}
	}
}

func (sampler *memoryPeakSampler) stopAndRead() uint64 {
	close(sampler.stop)
	<-sampler.done
	return sampler.peak.Load()
}
