package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sekolahkit/sekolah-app/internal/auth"
	"github.com/Sekolahkit/sekolah-app/internal/kelas"
	"github.com/Sekolahkit/sekolah-app/internal/laporan"
	"github.com/Sekolahkit/sekolah-app/internal/migration"
	"github.com/Sekolahkit/sekolah-app/internal/notifikasi"
	"github.com/Sekolahkit/sekolah-app/internal/payment"
	"github.com/Sekolahkit/sekolah-app/internal/pembayaran"
	"github.com/Sekolahkit/sekolah-app/internal/ppdb"
	"github.com/Sekolahkit/sekolah-app/internal/rekening"
	"github.com/Sekolahkit/sekolah-app/internal/sekolah"
	"github.com/Sekolahkit/sekolah-app/internal/setup"
	"github.com/Sekolahkit/sekolah-app/internal/siswa"
	"github.com/Sekolahkit/sekolah-app/internal/upload"
	"github.com/Sekolahkit/sekolah-app/pkg/frontend"
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

	workerCtx, workerCancel := context.WithCancel(context.Background())

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
				r.Post("/", pembayaranHandler.CreatePembayaran)
				r.Put("/{id}/verify", pembayaranHandler.VerifyPembayaran)
				r.Put("/{id}/reject", pembayaranHandler.RejectPembayaran)
			})
		})

		paymentRepo := payment.NewRepository(db)
		var gateways []payment.Gateway
		if secrets.MidtransServerKey != "" {
			gateways = append(gateways, payment.NewMidtrans(payment.MidtransConfig{
				ServerKey: secrets.MidtransServerKey,
				ClientKey: secrets.MidtransClientKey,
			}))
		}
		if secrets.XenditSecretKey != "" {
			gateways = append(gateways, payment.NewXendit(payment.XenditConfig{
				SecretKey:     secrets.XenditSecretKey,
				CallbackToken: secrets.XenditSecretKey,
			}))
		}
		paymentService := payment.NewService(paymentRepo, gateways...)
		paymentHandler := payment.NewHandler(paymentService)

		r.Post("/payment/callback/midtrans", paymentHandler.MidtransCallback)
		r.Post("/payment/callback/xendit", paymentHandler.XenditCallback)

		ppdbRepo := ppdb.NewRepository(db)
		ppdbService := ppdb.NewService(ppdbRepo)
		ppdbHandler := ppdb.NewHandler(ppdbService)

		r.Post("/ppdb/daftar", ppdbHandler.Daftar)
		r.Get("/ppdb/pengumuman/{id}", ppdbHandler.GetPengumuman)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))

			r.Get("/ppdb/pendaftar", ppdbHandler.ListPendaftar)
			r.Get("/ppdb/pendaftar/{id}", ppdbHandler.GetPendaftar)
			r.Put("/ppdb/pendaftar/{id}", ppdbHandler.UpdateStatus)
			r.Get("/ppdb/pendaftar/{id}/berkas", ppdbHandler.ListBerkas)
			r.Put("/ppdb/berkas/{id}", ppdbHandler.VerifikasiBerkas)
			r.Post("/ppdb/ujian", ppdbHandler.InputUjian)
			r.Post("/ppdb/pengumuman", ppdbHandler.PublishPengumuman)
			r.Get("/ppdb/konfigurasi-ranking", ppdbHandler.GetKonfigurasiRanking)
			r.Post("/ppdb/konfigurasi-ranking", ppdbHandler.UpsertKonfigurasiRanking)
		})

		notifRepo := notifikasi.NewRepository(db)
		notifService := notifikasi.NewService(notifRepo)
		notifHandler := notifikasi.NewHandler(notifService)

		registry := notifikasi.NewRegistry()
		if cfg.Notifikasi.WhatsApp {
			registry.Register(notifikasi.NewWhatsAppProvider())
		}
		if cfg.Notifikasi.Email {
			registry.Register(notifikasi.NewEmailProvider(notifikasi.EmailConfig{
				Host:     secrets.SMTPHost,
				Port:     secrets.SMTPPort,
				User:     secrets.SMTPUser,
				Password: secrets.SMTPPassword,
				From:     secrets.SMTPFrom,
			}))
		}
		if cfg.Notifikasi.Telegram {
			registry.Register(notifikasi.NewTelegramProvider(secrets.TelegramBotToken))
		}

		worker := notifikasi.NewWorker(notifRepo, registry, notifikasi.DefaultWorkerConfig())
		go worker.Start(workerCtx)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin"))

			r.Get("/notifikasi", notifHandler.ListNotifikasi)
			r.Post("/notifikasi/test", notifHandler.TestSend)
			r.Get("/notifikasi/queue", notifHandler.QueueStatus)
			r.Post("/notifikasi/{id}/retry", notifHandler.Retry)
		})

		laporanRepo := laporan.NewRepository(db)
		laporanService := laporan.NewService(laporanRepo)
		laporanHandler := laporan.NewHandler(laporanService)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin"))
			r.Get("/dashboard/admin", laporanHandler.DashboardAdmin)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))
			r.Get("/dashboard/operator", laporanHandler.DashboardOperator)
			r.Get("/laporan/pembayaran", laporanHandler.RekapPembayaran)
			r.Get("/laporan/ppdb", laporanHandler.RekapPPDB)
			r.Get("/laporan/siswa", laporanHandler.RekapSiswa)
		})

		uploadService := upload.NewService("./uploads", cfg.Upload.MaxSize, cfg.Upload.AllowedTypes)
		uploadHandler := upload.NewHandler(uploadService)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(secrets.JWTSecret))
			r.Post("/upload", uploadHandler.Upload)
			r.Get("/upload/*", uploadHandler.Download)
		})
	})

	if frontend.HasBuild() {
		r.NotFound(frontend.Handler().ServeHTTP)
		slog.Info("frontend embedded, serving SPA")
	} else {
		slog.Info("frontend not embedded, API-only mode")
	}

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	workerCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}

	slog.Info("server exited")
}
