package app

import (
	"fmt"
	"io"
	"time"
)

type vectorEmbeddingProgress struct {
	writer     io.Writer
	total      int
	completed  int
	started    time.Time
	lastReport time.Time
}

func newVectorEmbeddingProgress(writer io.Writer, total int) *vectorEmbeddingProgress {
	now := time.Now()
	return &vectorEmbeddingProgress{writer: writer, total: total, started: now, lastReport: now}
}

func (progress *vectorEmbeddingProgress) complete(count int) {
	progress.completed += count
	now := time.Now()
	if progress.completed < progress.total && now.Sub(progress.lastReport) < time.Second {
		return
	}
	elapsed := now.Sub(progress.started)
	rate := 0.0
	if elapsed > 0 {
		rate = float64(progress.completed) / elapsed.Seconds()
	}
	_, _ = fmt.Fprintf(
		progress.writer,
		"vector build: embedding %d/%d (%.2f vectors/s, %s elapsed)\n",
		progress.completed,
		progress.total,
		rate,
		formatVectorDuration(elapsed),
	)
	progress.lastReport = now
}

func formatVectorDuration(value time.Duration) string {
	if value < time.Second {
		return value.Round(time.Millisecond).String()
	}
	return value.Round(time.Second).String()
}
