-- +goose Up
-- +goose StatementBegin
CREATE TABLE wallets (
     id         UUID PRIMARY KEY,
     balance    BIGINT NOT NULL          DEFAULT 0,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE wallets;
-- +goose StatementEnd
