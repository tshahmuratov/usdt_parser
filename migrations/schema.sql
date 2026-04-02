CREATE TABLE rates (
    id BIGSERIAL PRIMARY KEY,
    ask DOUBLE PRECISION NOT NULL,
    bid DOUBLE PRECISION NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rates_fetched_at ON rates (fetched_at);
CREATE INDEX idx_rates_created_at ON rates (created_at);
