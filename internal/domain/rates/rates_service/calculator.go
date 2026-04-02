package rates_service

import "github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"

var (
	_ rates_model.CalcMethod = rates_model.TopN{}
	_ rates_model.CalcMethod = rates_model.AvgNM{}
)
