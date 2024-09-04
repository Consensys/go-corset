package util

import (
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// PerfStats provides a snapshot of memory allocation at a given point in time.
type PerfStats struct {
	// Starting time
	startTime time.Time
	// Starting total memory allocation
	startMem uint64
	// Starting number of gc events
	startGc uint32
}

// NewPerfStats creates a new snapshot of the current amount of memory allocated.
func NewPerfStats() *PerfStats {
	var m runtime.MemStats

	startTime := time.Now()

	runtime.ReadMemStats(&m)

	return &PerfStats{startTime, m.TotalAlloc, m.NumGC}
}

// Log logs the difference between the state now and as it was when the PerfStats object was created.
func (p *PerfStats) Log(prefix string) {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	alloc := (m.TotalAlloc - p.startMem) / 1024 / 1024 / 1024
	gcs := m.NumGC - p.startGc
	exectime := time.Since(p.startTime).Seconds()

	log.Debugf("%s took %0.2fs using %v Gb (%v GC events) [%v Gb]", prefix, exectime, alloc, gcs, m.Alloc/1024/1024/1024)
}
