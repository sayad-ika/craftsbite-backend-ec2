CREATE TABLE wfh_periods (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    start_date  DATE        NOT NULL,
    end_date    DATE        NOT NULL,
    reason      TEXT,
    created_by  UUID        NOT NULL REFERENCES users(id),
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wfh_period_dates CHECK (end_date >= start_date)
);
