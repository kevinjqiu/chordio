package chordio

import "github.com/kevinjqiu/chordio/telemetry"

type Config struct {
	M         Rank
	Bind      string
	Telemetry telemetry.Config
}
