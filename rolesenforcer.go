package sgul

// RolesEnforcer is the user roles enforcer for gain user access to resources.
type RolesEnforcer interface {
	Enforce(role string, route string, method string) bool
}

// MatchAllEnforcer is the match all rules;routes enforcer.
// It will always authorize a user.
type MatchAllEnforcer struct{}

// Enforce always authorize a user. It skips role/route/method checks.
func (mae *MatchAllEnforcer) Enforce(role string, route string, method string) bool {
	logger.Debugf("Enforcing user role for resource with MatchAllEnforcer strategy", "role", role, "route", route, "method", method)
	return true
}

// MatchRoleEnforcer authorize user only if it's role is in roles.
type MatchRoleEnforcer struct {
	roles []string
}

// NewMatchRoleEnforcer returns a new MathRoleEnforcer instance.
func NewMatchRoleEnforcer(roles []string) *MatchRoleEnforcer {
	return &MatchRoleEnforcer{roles: roles}
}

// Enforce always authorize a user role agains roles[]. It skips route/method checks.
// The use is authorized only if it role is in roles[].
// If roles[] size is 0 user will not be authorized.
func (mre *MatchRoleEnforcer) Enforce(role string, route string, method string) bool {
	logger.Debugf("Enforcing user role for resource with MatchRoleEnforcer strategy", "role", role, "route", route, "method", method)
	if len(mre.roles) > 0 {
		return ContainsString(mre.roles, role)
	}
	return false
}
