-- +goose Up
CREATE TABLE operations (
    id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    type TEXT NOT NULL,
    instrument_type TEXT NOT NULL,
    figi TEXT NOT NULL,
    ticker TEXT NOT NULL,
    quantity TEXT NOT NULL,
    payment_currency TEXT NOT NULL,
    payment_units TEXT NOT NULL,
    payment_nano INTEGER NOT NULL,
    PRIMARY KEY (id, account_id)
);

-- +goose Down
DROP TABLE operations;
