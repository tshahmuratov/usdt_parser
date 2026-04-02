package rates_handler

import (
	"context"
	"errors"

	ratesv1 "github.com/tshahmuratov/usdt_parser/gen/rates/v1"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateServicer interface {
	GetRates(ctx context.Context, method rates_model.CalcMethod) (*rates_model.Rate, error)
}

type RatesHandler struct {
	ratesv1.UnimplementedRateServiceServer
	svc RateServicer
}

func NewRatesHandler(svc RateServicer) *RatesHandler {
	return &RatesHandler{svc: svc}
}

func (h *RatesHandler) GetRates(ctx context.Context, req *ratesv1.GetRatesRequest) (*ratesv1.GetRatesResponse, error) {
	method, err := toCalcMethod(req.GetMethod())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid calc method: %v", err)
	}

	rate, err := h.svc.GetRates(ctx, method)
	if err != nil {
		return nil, mapError(err)
	}

	return toGetRatesResponse(rate), nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, rates_model.ErrIndexOutOfBounds),
		errors.Is(err, rates_model.ErrInvalidRange),
		errors.Is(err, rates_model.ErrEmptyEntries):
		return status.Errorf(codes.InvalidArgument, "%v", err)
	case errors.Is(err, rates_model.ErrFetchFailed):
		return status.Errorf(codes.Unavailable, "%v", err)
	case errors.Is(err, rates_model.ErrStoreFailed):
		return status.Errorf(codes.Internal, "%v", err)
	default:
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}
}
