package grinex

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

var _ rates_interface.ExchangeClient = (*GrinexClient)(nil)

type depthResponse struct {
	Timestamp int64        `json:"timestamp"`
	Asks      []depthEntry `json:"asks"`
	Bids      []depthEntry `json:"bids"`
}

type depthEntry struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
}

type GrinexClient struct {
	client  *resty.Client
	baseURL string
}

func NewGrinexClient(baseURL string, timeout time.Duration) *GrinexClient {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	c := resty.New().SetTimeout(timeout)
	return &GrinexClient{client: c, baseURL: baseURL}
}

func (g *GrinexClient) FetchDepth(ctx context.Context) (*rates_model.SpotDepth, error) {
	r, err := g.client.R().
		SetContext(ctx).
		SetQueryParam("symbol", "usdta7a5").
		Get(g.baseURL + "/api/v1/spot/depth")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rates_model.ErrFetchFailed, err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("%w: status %d", rates_model.ErrFetchFailed, r.StatusCode())
	}

	var resp depthResponse
	if err := json.Unmarshal(r.Body(), &resp); err != nil {
		return nil, fmt.Errorf("%w: %v", rates_model.ErrFetchFailed, err)
	}

	depth := &rates_model.SpotDepth{
		Timestamp: time.Unix(resp.Timestamp, 0).UTC(),
		Asks:      make([]rates_model.SpotEntry, 0, len(resp.Asks)),
		Bids:      make([]rates_model.SpotEntry, 0, len(resp.Bids)),
	}

	for _, e := range resp.Asks {
		entry, err := parseEntry(e)
		if err != nil {
			return nil, fmt.Errorf("%w: parse ask: %v", rates_model.ErrFetchFailed, err)
		}
		depth.Asks = append(depth.Asks, entry)
	}
	for _, e := range resp.Bids {
		entry, err := parseEntry(e)
		if err != nil {
			return nil, fmt.Errorf("%w: parse bid: %v", rates_model.ErrFetchFailed, err)
		}
		depth.Bids = append(depth.Bids, entry)
	}

	return depth, nil
}

func parseEntry(e depthEntry) (rates_model.SpotEntry, error) {
	price, err := strconv.ParseFloat(e.Price, 64)
	if err != nil {
		return rates_model.SpotEntry{}, fmt.Errorf("parse price %q: %w", e.Price, err)
	}
	volume, err := strconv.ParseFloat(e.Volume, 64)
	if err != nil {
		return rates_model.SpotEntry{}, fmt.Errorf("parse volume %q: %w", e.Volume, err)
	}
	amount, err := strconv.ParseFloat(e.Amount, 64)
	if err != nil {
		return rates_model.SpotEntry{}, fmt.Errorf("parse amount %q: %w", e.Amount, err)
	}
	return rates_model.SpotEntry{
		Price:  rates_model.Price(price),
		Volume: volume,
		Amount: amount,
	}, nil
}
