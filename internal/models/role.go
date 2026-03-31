package models

// Role represents user roles in the system
type Role string

const (
	RoleEmployee  Role = "employee"
	RoleTeamLead  Role = "team_lead"
	RoleAdmin     Role = "admin"
	RoleLogistics Role = "logistics"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleEmployee, RoleTeamLead, RoleAdmin, RoleLogistics:
		return true
	}
	return false
}

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}
