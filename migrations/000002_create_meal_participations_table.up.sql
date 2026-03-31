CREATE TABLE meal_participations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL,
    is_participating BOOLEAN NOT NULL DEFAULT true,
    opted_out_at TIMESTAMP WITH TIME ZONE,
    override_by UUID REFERENCES users(id),
    override_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_meal_participations_user_date ON meal_participations(user_id, date);
CREATE INDEX idx_meal_participations_date ON meal_participations(date);
CREATE UNIQUE INDEX unique_user_date_meal ON meal_participations(user_id, date, meal_type);
