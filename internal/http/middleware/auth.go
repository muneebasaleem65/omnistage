package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/muneebasaleem65/omnistage/internal/token"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			tokenString, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || tokenString == "" {
				response.WriteJson(w, http.StatusUnauthorized,
					response.GeneralError(fmt.Errorf("missing or malformed token")))
				return
			}

			claims, err := token.Parse(tokenString, secret)
			if err != nil {
				response.WriteJson(w, http.StatusUnauthorized,
					response.GeneralError(fmt.Errorf("invalid token")))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}
