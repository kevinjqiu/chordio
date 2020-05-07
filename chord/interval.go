package chord

import (
	"math"
)

type IntervalOption func(i *Interval)

const (
	intervalOptionOpen int = iota
	intervalOptionClosed
)

type Interval struct {
	m           Rank
	start, end  ChordID
	leftOption  int
	rightOption int
}

func WithLeftOpen(i *Interval) {
	i.leftOption = intervalOptionOpen
}

func WithLeftClosed(i *Interval) {
	i.leftOption = intervalOptionClosed
}

func WithRightOpen(i *Interval) {
	i.rightOption = intervalOptionOpen
}

func WithRightClosed(i *Interval) {
	i.rightOption = intervalOptionClosed
}

func (i Interval) Has(id ChordID) bool {
	var leftClause, rightClause bool

	if i.leftOption == intervalOptionOpen {
		leftClause = i.start < id
	} else {
		leftClause = i.start <= id
	}

	if i.rightOption == intervalOptionOpen {
		rightClause = id < i.end
	} else {
		rightClause = id <= i.end
	}

	if i.start < i.end {
		return leftClause && rightClause
	}
	max := ChordID(pow2(uint32(i.m)))
	return leftClause && id < max || 0 <= id && rightClause
}

func NewInterval(m Rank, start, end ChordID, options ...IntervalOption) Interval {
	i := Interval{
		m:           m,
		start:       start,
		end:         end,
		leftOption:  intervalOptionClosed,
		rightOption: intervalOptionOpen,
	}

	for _, opt := range options {
		opt(&i)
	}

	return i
}

func pow2(exp uint32) uint64 {
	return uint64(math.Pow(2, float64(exp)))
}
