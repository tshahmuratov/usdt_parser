package rates_model

import "time"

type Rate struct {
	ID        int64
	Ask       Price
	Bid       Price
	FetchedAt time.Time
	CreatedAt time.Time
}
