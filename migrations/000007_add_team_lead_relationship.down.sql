-- Drop indexes first
DROP INDEX IF EXISTS idx_team_members_user_id;
DROP INDEX IF EXISTS idx_team_members_team_id;
DROP INDEX IF EXISTS idx_teams_active;
DROP INDEX IF EXISTS idx_teams_team_lead_id;

-- Drop tables (team_members first due to foreign key constraint)
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
