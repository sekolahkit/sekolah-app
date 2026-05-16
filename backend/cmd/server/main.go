package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Sekolahkit/sekolah-app/internal/auth"
	"github.com/Sekolahkit/sekolah-app/internal/kelas"
	"github.com/Sekolahkit/sekolah-app/internal/migration"
	"github.com/Sekolahkit/sekolah-app/internal/pembayaran"
	"github.com/Sekolahkit/sekolah-app/internal/rekening"
	"github.com/Sekolahkit/sekolah-app/internal/sekolah"
	"github.com/Sekolahkit/sekolah-app/internal/setup"
	"github.com/Sekolahkit/sekolah-app/internal/siswa"
	"github.com/Sekolahkit/sekolah-app/pkg/config"
	"github.com/Sekolahkit/sekolah-app/pkg/database"
	mw "github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	secrets := config.LoadSecrets()
	if err := config.ValidateSecrets(secrets); err != nil {
		slog.Error("invalid secrets", "error", err)
		os.Exit(1)
	}

	slog.Info("starting",
		"app", cfg.App.Name,
		"port", cfg.App.Port,
		"database", cfg.Database.Path,
	)

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migration.Run(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations up-to-date")

	r := chi.NewRouter()

	r.Use(mw.Recover)
	r.Use(mw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.RateLimit(cfg.RateLimit.API))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok","version":"0.1.0","database":"connected"}`)
	})

	r.Route("/api/v1", func(r chi.Router) {
		setupRepo := setup.NewRepository(db)
		setupService := setup.NewService(setupRepo)
		setupHandler := setup.NewHandler(setupService)

		r.Route("/setup", func(r chi.Router) {
			r.Use(setupHandler.SetupGuard)
			r.Get("/status", setupHandler.Status)
			r.Post("/", setupHandler.RunSetup)
		})

		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, secrets.JWTSecret)
		authHandler := auth.NewHandler(authService)

		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/auth/me", authHandler.Me)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin"))
			r.Post("/auth/revoke-all/{user_id}", authHandler.RevokeAll)
		})

		sekolahRepo := sekolah.NewRepository(db)
		sekolahService := sekolah.NewService(sekolahRepo)
		sekolahHandler := sekolah.NewHandler(sekolahService)

		siswaRepo := siswa.NewRepository(db)
		siswaService := siswa.NewService(siswaRepo)
		siswaHandler := siswa.NewHandler(siswaService)

		kelasRepo := kelas.NewRepository(db)
		kelasService := kelas.NewService(kelasRepo)
		kelasHandler := kelas.NewHandler(kelasService)

		rekeningRepo := rekening.NewRepository(db)
		rekeningService := rekening.NewService(rekeningRepo)
		rekeningHandler := rekening.NewHandler(rekeningService)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))

			r.Get("/sekolah", sekolahHandler.Get)

			r.Get("/rekening-sekolah/aktif", rekeningHandler.ListAktif)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin"))

			r.Put("/sekolah", sekolahHandler.Update)

			r.Route("/rekening-sekolah", func(r chi.Router) {
				r.Get("/", rekeningHandler.List)
				r.Post("/", rekeningHandler.Create)
				r.Put("/{id}", rekeningHandler.Update)
				r.Delete("/{id}", rekeningHandler.Delete)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))

			r.Route("/siswa", func(r chi.Router) {
				r.Get("/", siswaHandler.List)
				r.Get("/{id}", siswaHandler.GetByID)
				r.Post("/", siswaHandler.Create)
				r.Put("/{id}", siswaHandler.Update)
				r.Delete("/{id}", siswaHandler.Delete)
				r.Post("/import", siswaHandler.Import)
				r.Get("/export", siswaHandler.Export)
			})

			r.Route("/kelas", func(r chi.Router) {
				r.Get("/", kelasHandler.List)
				r.Get("/{id}", kelasHandler.GetByID)
				r.Post("/", kelasHandler.Create)
				r.Put("/{id}", kelasHandler.Update)
				r.Delete("/{id}", kelasHandler.Delete)
				r.Get("/{id}/siswa", kelasHandler.ListSiswa)
				r.Post("/{id}/siswa", kelasHandler.AddSiswa)
				r.Delete("/{id}/siswa/{siswa_id}", kelasHandler.RemoveSiswa)
			})
		})

		pembayaranRepo := pembayaran.NewRepository(db)
		pembayaranService := pembayaran.NewService(pembayaranRepo)
		pembayaranHandler := pembayaran.NewHandler(pembayaranService)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))

			r.Route("/kategori-pembayaran", func(r chi.Router) {
				r.Get("/", pembayaranHandler.ListKategori)
				r.Post("/", pembayaranHandler.CreateKategori)
				r.Put("/{id}", pembayaranHandler.UpdateKategori)
				r.Delete("/{id}", pembayaranHandler.DeleteKategori)
			})

			r.Route("/tagihan", func(r chi.Router) {
				r.Get("/", pembayaranHandler.ListTagihan)
				r.Get("/{id}", pembayaranHandler.GetTagihanByID)
				r.Post("/", pembayaranHandler.CreateTagihan)
				r.Post("/bulk", pembayaranHandler.BulkCreateTagihan)
				r.Put("/{id}", pembayaranHandler.UpdateTagihan)
				r.Delete("/{id}", pembayaranHandler.DeleteTagihan)
			})

			r.Route("/pembayaran", func(r chi.Router) {
				r.Get("/", pembayaranHandler.ListPembayaran)
				r.Get("/{id}", pembayaranHandler.GetPembayaranByID)
				r.Put("/{id}/verify", pembayaranHandler.VerifyPembayaran)
				r.Put("/{id}/reject", pembayaranHandler.RejectPembayaran)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Post("/pembayaran", pembayaranHandler.CreatePembayaran)
		})
	})

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	slog.Info("server listening", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
