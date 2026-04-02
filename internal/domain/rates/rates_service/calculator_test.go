package rates_service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

func entries(prices ...float64) []rates_model.SpotEntry {
	out := make([]rates_model.SpotEntry, len(prices))
	for i, p := range prices {
		out[i] = rates_model.SpotEntry{Price: rates_model.Price(p)}
	}
	return out
}

func TestTopN_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		entries []rates_model.SpotEntry
		want    rates_model.Price
		wantErr error
	}{
		{"first element", 0, entries(10, 20, 30), 10, nil},
		{"middle element", 1, entries(10, 20, 30), 20, nil},
		{"last element", 2, entries(10, 20, 30), 30, nil},
		{"empty entries", 0, entries(), 0, rates_model.ErrEmptyEntries},
		{"out of bounds", 5, entries(10, 20), 0, rates_model.ErrIndexOutOfBounds},
		{"negative index", -1, entries(10), 0, rates_model.ErrIndexOutOfBounds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := rates_model.TopN{N: tt.n}
			got, err := calc.Calculate(tt.entries)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAvgNM_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		n, m    int
		entries []rates_model.SpotEntry
		want    rates_model.Price
		wantErr error
	}{
		{"single element", 0, 0, entries(10), 10, nil},
		{"full range", 0, 2, entries(10, 20, 30), 20, nil},
		{"sub range", 1, 2, entries(10, 20, 30), 25, nil},
		{"empty entries", 0, 0, entries(), 0, rates_model.ErrEmptyEntries},
		{"n > m", 2, 1, entries(10, 20, 30), 0, rates_model.ErrInvalidRange},
		{"m out of bounds", 0, 5, entries(10, 20), 0, rates_model.ErrIndexOutOfBounds},
		{"n negative", -1, 0, entries(10), 0, rates_model.ErrIndexOutOfBounds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := rates_model.AvgNM{N: tt.n, M: tt.m}
			got, err := calc.Calculate(tt.entries)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
