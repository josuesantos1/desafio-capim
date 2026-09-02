CREATE TABLE dentists (
    id          uuid PRIMARY KEY,
    clinic_id   uuid NOT NULL REFERENCES clinics(id),
    name        text NOT NULL,
    phone       text NOT NULL,
    email       text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX dentists_clinic_email_active_uidx
    ON dentists (clinic_id, email)
    WHERE deleted_at IS NULL;

CREATE INDEX dentists_clinic_id_idx
    ON dentists (clinic_id)
    WHERE deleted_at IS NULL;
