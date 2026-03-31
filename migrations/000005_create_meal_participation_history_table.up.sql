CREATE TABLE meal_participation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    previous_value VARCHAR(20),
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_history_user_date ON meal_participation_history(user_id, date);
CREATE INDEX idx_history_created_at ON meal_participation_history(created_at);
