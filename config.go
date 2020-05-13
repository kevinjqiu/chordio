package chordio

import (
	"github.com/kevinjqiu/chordio/chord"
	"time"
)

type StabilizationConfig struct {
	Disabled bool
	Period   time.Duration
	Jitter   time.Duration
}

type Config struct {
	ID   chord.ID
	M    chord.Rank
	Bind string
	// Disable the stabilization protocol for debugging purposes
	Stabilization StabilizationConfig
}
