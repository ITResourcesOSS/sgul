package sgul

import (
	"context"
	"fmt"

	"github.com/casbin/casbin"

	"github.com/go-chi/chi/middleware"
)

// RolesEnforcer is the user roles enforcer for gain user access to resources.
type RolesEnforcer interface {
	Enforce(ctx context.Context, role string, route string, method string) bool
}

// MatchAllEnforcer is the match all rules;routes enforcer.
// It will always authorize a user.
type MatchAllEnforcer struct{}

// Enforce always authorize a user. It skips role/route/method checks.
func (mae *MatchAllEnforcer) Enforce(ctx context.Context, role string, route string, method string) bool {
	logEnforce(ctx, role, route, method, "MatchAllEnforcer")
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
func (mre *MatchRoleEnforcer) Enforce(ctx context.Context, role string, route string, method string) bool {
	logEnforce(ctx, role, route, method, "MatchRoleEnforcer")
	if len(mre.roles) > 0 {
		return ContainsString(mre.roles, role)
	}
	return false
}

// CasbinEnforcer .
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewCasbinEnforcer .
func NewCasbinEnforcer() *CasbinEnforcer {
	return &CasbinEnforcer{
		enforcer: casbin.NewEnforcer("./auth_model.conf", "./policy.csv"),
	}
}

// Enforce proxy casbin Enforce func.
func (ce *CasbinEnforcer) Enforce(ctx context.Context, role string, route string, method string) bool {
	logEnforce(ctx, role, route, method, "CasbinEnforcer")
	res, err := ce.enforcer.EnforceSafe(role, route, method)
	if err != nil {
		logger.Errorw("unable to enforce with casbin enforcer", "error", err, "request-id", middleware.GetReqID(ctx))
		return false
	}

	if res {
		return true
	}

	return false
}

func logEnforce(ctx context.Context, role string, route string, method string, kind string) {
	logger.Infow(fmt.Sprintf("Enforcing role '%s' for '%s %s' with %s strategy", role, method, route, kind),
		"request-id", middleware.GetReqID(ctx))
}
