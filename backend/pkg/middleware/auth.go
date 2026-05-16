package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	UserID    int64  `json:"user_id"`
	SekolahID int64  `json:"sekolah_id"`
	Role      string `json:"role"`
	Email     string `json:"email"`
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				response.Error(w, 401, "UNAUTHORIZED", "Belum login")
				return
			}

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}
			sekolahIDFloat, ok := claims["sekolah_id"].(float64)
			if !ok {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}
			role, ok := claims["role"].(string)
			if !ok {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}
			email, ok := claims["email"].(string)
			if !ok {
				response.Error(w, 401, "UNAUTHORIZED", "Token tidak valid")
				return
			}

			user := &UserClaims{
				UserID:    int64(userIDFloat),
				SekolahID: int64(sekolahIDFloat),
				Role:      role,
				Email:     email,
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				response.Error(w, 401, "UNAUTHORIZED", "Belum login")
				return
			}

			for _, role := range roles {
				if strings.EqualFold(user.Role, role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, 403, "FORBIDDEN", "Tidak punya akses")
		})
	}
}

func GetUser(ctx context.Context) *UserClaims {
	user, _ := ctx.Value(UserContextKey).(*UserClaims)
	return user
}
