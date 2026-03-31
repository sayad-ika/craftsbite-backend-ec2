CREATE TABLE work_location_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    location VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    previous_location VARCHAR(20),
    override_by UUID REFERENCES users(id) ON DELETE SET NULL,
    override_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
