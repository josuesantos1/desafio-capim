CREATE TABLE payments (
    id            uuid PRIMARY KEY,
    clinic_id     uuid NOT NULL REFERENCES clinics(id),
    dentist_id    uuid REFERENCES dentists(id),
    amount_cents  bigint NOT NULL CHECK (amount_cents > 0),
    status        text NOT NULL CHECK (status IN ('pending', 'approved')),
    pix_code      text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    approved_at   timestamptz
);

CREATE INDEX payments_clinic_id_idx ON payments (clinic_id);
