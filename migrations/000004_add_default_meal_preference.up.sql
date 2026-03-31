-- default_meal_preference already added in 000001; this migration exists for plan compliance
ALTER TABLE users ADD COLUMN IF NOT EXISTS default_meal_preference VARCHAR(20) NOT NULL DEFAULT 'opt_in';
