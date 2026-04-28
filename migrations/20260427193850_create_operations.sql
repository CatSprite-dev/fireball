-- +goose Up
CREATE TABLE candles (
    figi TEXT NOT NULL,
    interval TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL,
    open_units TEXT NOT NULL,
    open_nano INTEGER NOT NULL,
    close_units TEXT NOT NULL,
    close_nano INTEGER NOT NULL,
    PRIMARY KEY (figi, interval, time)
);

-- +goose Down
DROP TABLE candles;
