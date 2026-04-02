package rates_handler

import (
	"fmt"

	ratesv1 "github.com/tshahmuratov/usdt_parser/gen/rates/v1"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toCalcMethod(pb *ratesv1.CalcMethod) (rates_model.CalcMethod, error) {
	if pb == nil {
		return nil, fmt.Errorf("calc method is required")
	}
	switch m := pb.Method.(type) {
	case *ratesv1.CalcMethod_TopN:
		return rates_model.TopN{N: int(m.TopN.N)}, nil
	case *ratesv1.CalcMethod_AvgNm:
		return rates_model.AvgNM{N: int(m.AvgNm.N), M: int(m.AvgNm.M)}, nil
	default:
		return nil, fmt.Errorf("unknown calc method type")
	}
}

func toGetRatesResponse(rate *rates_model.Rate) *ratesv1.GetRatesResponse {
	return &ratesv1.GetRatesResponse{
		Ask:       float64(rate.Ask),
		Bid:       float64(rate.Bid),
		Timestamp: timestamppb.New(rate.FetchedAt),
	}
}
