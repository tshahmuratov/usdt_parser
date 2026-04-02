package rates_model

import (
	"errors"
	"time"
)

var (
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrInvalidRange     = errors.New("invalid range: n > m")
	ErrEmptyEntries     = errors.New("empty entries")
	ErrStoreFailed      = errors.New("failed to store rate")
	ErrFetchFailed      = errors.New("failed to fetch depth")
)

type Price float64

type SpotEntry struct {
	Price  Price
	Volume float64
	Amount float64
}

type SpotDepth struct {
	Asks      []SpotEntry
	Bids      []SpotEntry
	Timestamp time.Time
}

type CalcMethod interface {
	Calculate(entries []SpotEntry) (Price, error)
}

type TopN struct {
	N int
}

type AvgNM struct {
	N int
	M int
}

func (t TopN) Calculate(entries []SpotEntry) (Price, error) {
	if len(entries) == 0 {
		return 0, ErrEmptyEntries
	}
	if t.N < 0 || t.N >= len(entries) {
		return 0, ErrIndexOutOfBounds
	}
	return entries[t.N].Price, nil
}

func (a AvgNM) Calculate(entries []SpotEntry) (Price, error) {
	if len(entries) == 0 {
		return 0, ErrEmptyEntries
	}
	if a.N < 0 || a.M < 0 {
		return 0, ErrIndexOutOfBounds
	}
	if a.N > a.M {
		return 0, ErrInvalidRange
	}
	if a.M >= len(entries) {
		return 0, ErrIndexOutOfBounds
	}
	var sum Price
	for i := a.N; i <= a.M; i++ {
		sum += entries[i].Price
	}
	return sum / Price(a.M-a.N+1), nil
}
