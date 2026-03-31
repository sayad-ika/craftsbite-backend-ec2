ALTER TABLE bulk_opt_outs
    DROP COLUMN IF EXISTS override_by,
    DROP COLUMN IF EXISTS override_reason;
