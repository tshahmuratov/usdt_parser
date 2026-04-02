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
	mocks "github.com/tshahmuratov/usdt_parser/mocks/domain/rates/rates_handler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRatesHandler_GetRates(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		svc := mocks.NewMockRateServicer(t)
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
		assert.NotNil(t, resp.FetchedAt)
		assert.Equal(t, now.Unix(), resp.FetchedAt.AsTime().Unix())
	})

	t.Run("nil method", func(t *testing.T) {
		svc := mocks.NewMockRateServicer(t)
		handler := rates_handler.NewRatesHandler(svc)

		_, err := handler.GetRates(context.Background(), &ratesv1.GetRatesRequest{})
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("index out of bounds", func(t *testing.T) {
		svc := mocks.NewMockRateServicer(t)
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
