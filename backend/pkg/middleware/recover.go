package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Sekolahkit/sekolah-app/pkg/response"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"stack", string(debug.Stack()),
				)
				response.Error(w, 500, "INTERNAL_ERROR", "Terjadi kesalahan internal")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
