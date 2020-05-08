package chord

import (
	"bytes"
	"fmt"
	"math"
)

type IntervalOption func(i *Interval)

const (
	intervalOptionOpen int = iota
	intervalOptionClosed
)

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

type Interval struct {
	m           Rank
	Start, End  ID
	leftOption  int
	rightOption int
}

func (i Interval) String() string {
	var b bytes.Buffer

	switch i.leftOption {
	case intervalOptionOpen:
		b.WriteString("(")
	case intervalOptionClosed:
		b.WriteString("[")
	}
	b.WriteString(fmt.Sprintf("%d, %d", i.Start, i.End))
	switch i.rightOption {
	case intervalOptionOpen:
		b.WriteString(")")
	case intervalOptionClosed:
		b.WriteString("]")
	}
	return b.String()
}

func (i Interval) Has(id ID) bool {
	var leftClause, rightClause bool

	if i.leftOption == intervalOptionOpen {
		leftClause = i.Start < id
	} else {
		leftClause = i.Start <= id
	}

	if i.rightOption == intervalOptionOpen {
		rightClause = id < i.End
	} else {
		rightClause = id <= i.End
	}

	if i.Start < i.End {
		return leftClause && rightClause
	}
	max := ID(pow2(uint32(i.m)))
	return leftClause && id < max || 0 <= id && rightClause
}

func NewInterval(m Rank, start, end ID, options ...IntervalOption) Interval {
	i := Interval{
		m:           m,
		Start:       start,
		End:         end,
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
