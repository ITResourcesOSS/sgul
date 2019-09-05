package sgul

import (
	"context"
	"errors"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// Principal defines the struct registered into the Context
// representing the authenticated user information from the JWT Token.
type Principal struct {
	Username string
	Role     string
}

type ctxKey int

const ctxPrincipalKey ctxKey = iota

// ErrPrincipalNotInContext is returned if there is no Principal in the request context.
var ErrPrincipalNotInContext = errors.New("No Principal in request context")

// JwtAuthenticator is the JWT authentication middleware.
func JwtAuthenticator(roles []string) func(next http.Handler) http.Handler {
	conf := GetConfiguration().API.Security
	secret := []byte(conf.Jwt.Secret)

	jwtAuthenticator := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			trimmedAuth := strings.Fields(authorization)

			// Trim out Bearer from Authorization Header
			if authorization == "" || len(trimmedAuth) == 0 {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(trimmedAuth[1], claims,
				func(token *jwt.Token) (interface{}, error) {
					return secret, nil
				})
			if err != nil {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			principal := Principal{
				Username: claims["sub"].(string),
				Role:     claims["auth"].(string),
			}

			// check roles authorization: 403 Forbidden iff check fails
			if !ContainsString(roles, principal.Role) {
				http.Error(w, "", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), ctxPrincipalKey, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
	return jwtAuthenticator
}

// GetPrincipal return the user authenticated Princiapl information from the request context.
func GetPrincipal(ctx context.Context) (Principal, error) {
	if principal, ok := ctx.Value(ctxPrincipalKey).(Principal); ok {
		return principal, nil
	}
	return Principal{}, ErrPrincipalNotInContext
}
