package rates_handler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ratesv1 "github.com/tshahmuratov/usdt_parser/gen/rates/v1"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_handler"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockRateService struct {
	mock.Mock
}

func (m *mockRateService) GetRates(ctx context.Context, method rates_model.CalcMethod) (*rates_model.Rate, error) {
	args := m.Called(ctx, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rates_model.Rate), args.Error(1)
}

func TestRatesHandler_GetRates(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		svc := new(mockRateService)
		handler := rates_handler.NewRatesHandler(svc)

		svc.On("GetRates", mock.Anything, rates_model.TopN{N: 0}).Return(&rates_model.Rate{
			ID: 1, Ask: 81.24, Bid: 81.17, FetchedAt: now,
		}, nil)

		resp, err := handler.GetRates(context.Background(), &ratesv1.GetRatesRequest{
			Method: &ratesv1.CalcMethod{
				Method: &ratesv1.CalcMethod_TopN{TopN: &ratesv1.TopN{N: 0}},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 81.24, resp.Ask)
		assert.Equal(t, 81.17, resp.Bid)
	})

	t.Run("nil method", func(t *testing.T) {
		svc := new(mockRateService)
		handler := rates_handler.NewRatesHandler(svc)

		_, err := handler.GetRates(context.Background(), &ratesv1.GetRatesRequest{})
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("index out of bounds", func(t *testing.T) {
		svc := new(mockRateService)
		handler := rates_handler.NewRatesHandler(svc)

		svc.On("GetRates", mock.Anything, mock.Anything).Return(nil, rates_model.ErrIndexOutOfBounds)

		_, err := handler.GetRates(context.Background(), &ratesv1.GetRatesRequest{
			Method: &ratesv1.CalcMethod{
				Method: &ratesv1.CalcMethod_TopN{TopN: &ratesv1.TopN{N: 99}},
			},
		})
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
