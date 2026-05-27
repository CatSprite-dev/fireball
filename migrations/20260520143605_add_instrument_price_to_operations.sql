-- +goose Up
ALTER TABLE operations ADD COLUMN instrument_price_units TEXT NOT NULL DEFAULT '0';
ALTER TABLE operations ADD COLUMN instrument_price_nano INTEGER NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN instrument_price_currency TEXT NOT NULL DEFAULT 'RUB';
ALTER TABLE operations ADD COLUMN position_uid TEXT NOT NULL;

-- +goose Down
ALTER TABLE operations DROP COLUMN instrument_price_units;
ALTER TABLE operations DROP COLUMN instrument_price_nano;
ALTER TABLE operations DROP COLUMN instrument_price_currency;
ALTER TABLE operations DROP COLUMN position_uid;