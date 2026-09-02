CREATE TABLE clinics (
    id          uuid PRIMARY KEY,
    document    varchar(14) NOT NULL,
    legal_name  text NOT NULL,
    trade_name  text NOT NULL,
    bank        text,
    agency      text,
    account     text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX clinics_document_active_uidx
    ON clinics (document)
    WHERE deleted_at IS NULL;
