-- +goose Up
CREATE TABLE keyvals (
    key TEXT PRIMARY KEY,
    val TEXT NOT NULL,
    created_at TIMESTAMPZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPZ NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE FUNCTION update_keyval_updated_at()
    RETURNS trigger AS $$
    BEGIN
        NEW.updated_at = now();
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER keval_updated_at AFTER UPDATE
    ON keyvals
    FOR EACH ROW
    EXECUTE FUNCTION update_keyval_updated_at();

-- +goose Down
DROP TRIGGER keyval_updated_at;
DROP TABLE keyvals;
