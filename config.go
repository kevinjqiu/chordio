package chordio

import "github.com/kevinjqiu/chordio/chord"

type Config struct {
	ID   chord.ID
	M    chord.Rank
	Bind string
}
