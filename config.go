package chordio

import "github.com/kevinjqiu/chordio/chord"

type Config struct {
	ID   chord.ChordID
	M    chord.Rank
	Bind string
}
