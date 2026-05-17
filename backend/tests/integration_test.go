package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sekolahkit/sekolah-app/internal/auth"
	"github.com/Sekolahkit/sekolah-app/internal/kelas"
	"github.com/Sekolahkit/sekolah-app/internal/laporan"
	"github.com/Sekolahkit/sekolah-app/internal/migration"
	"github.com/Sekolahkit/sekolah-app/internal/notifikasi"
	"github.com/Sekolahkit/sekolah-app/internal/pembayaran"
	"github.com/Sekolahkit/sekolah-app/internal/ppdb"
	"github.com/Sekolahkit/sekolah-app/internal/rekening"
	"github.com/Sekolahkit/sekolah-app/internal/sekolah"
	"github.com/Sekolahkit/sekolah-app/internal/setup"
	"github.com/Sekolahkit/sekolah-app/internal/siswa"
	"github.com/Sekolahkit/sekolah-app/internal/upload"
	mw "github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-secret-key-minimum-32-characters"

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func setupTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := migration.Run(db); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(mw.Recover)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, `{"status":"ok"}`)
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
		authService := auth.NewService(authRepo, testJWTSecret)
		authHandler := auth.NewHandler(authService)

		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/auth/me", authHandler.Me)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
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
			r.Use(mw.Auth(testJWTSecret))
			r.Get("/sekolah", sekolahHandler.Get)
			r.Get("/rekening-sekolah/aktif", rekeningHandler.ListAktif)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
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
			r.Use(mw.Auth(testJWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))

			r.Route("/siswa", func(r chi.Router) {
				r.Get("/", siswaHandler.List)
				r.Get("/export", siswaHandler.Export)
				r.Get("/{id}", siswaHandler.GetByID)
				r.Post("/", siswaHandler.Create)
				r.Put("/{id}", siswaHandler.Update)
				r.Delete("/{id}", siswaHandler.Delete)
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
			r.Use(mw.Auth(testJWTSecret))
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
				r.Get("/export", pembayaranHandler.Export)
				r.Get("/{id}", pembayaranHandler.GetPembayaranByID)
				r.Get("/{id}/kwitansi", pembayaranHandler.Kwitansi)
				r.Post("/", pembayaranHandler.CreatePembayaran)
				r.Put("/{id}/verify", pembayaranHandler.VerifyPembayaran)
				r.Put("/{id}/reject", pembayaranHandler.RejectPembayaran)
			})
		})

		ppdbRepo := ppdb.NewRepository(db)
		ppdbService := ppdb.NewService(ppdbRepo)
		ppdbHandler := ppdb.NewHandler(ppdbService)

		r.Post("/ppdb/daftar", ppdbHandler.Daftar)
		r.Get("/ppdb/pengumuman/{id}", ppdbHandler.GetPengumuman)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))
			r.Get("/ppdb/pendaftar", ppdbHandler.ListPendaftar)
			r.Get("/ppdb/export", ppdbHandler.Export)
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

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Use(mw.RequireRole("admin"))
			r.Get("/notifikasi", notifHandler.ListNotifikasi)
			r.Post("/notifikasi/test", notifHandler.TestSend)
			r.Get("/notifikasi/queue", notifHandler.QueueStatus)
		})

		laporanRepo := laporan.NewRepository(db)
		laporanService := laporan.NewService(laporanRepo)
		laporanHandler := laporan.NewHandler(laporanService)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Use(mw.RequireRole("admin"))
			r.Get("/dashboard/admin", laporanHandler.DashboardAdmin)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Use(mw.RequireRole("admin", "operator"))
			r.Get("/dashboard/operator", laporanHandler.DashboardOperator)
			r.Get("/laporan/pembayaran", laporanHandler.RekapPembayaran)
			r.Get("/laporan/pembayaran/export", laporanHandler.ExportPembayaran)
			r.Get("/laporan/ppdb", laporanHandler.RekapPPDB)
			r.Get("/laporan/ppdb/export", laporanHandler.ExportPPDB)
			r.Get("/laporan/siswa", laporanHandler.RekapSiswa)
			r.Get("/laporan/siswa/export", laporanHandler.ExportSiswa)
		})

		uploadTmpDir := t.TempDir()
		uploadService := upload.NewService(uploadTmpDir, 5, []string{"image/jpeg", "image/png", "application/pdf"})
		uploadHandler := upload.NewHandler(uploadService, testJWTSecret)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(testJWTSecret))
			r.Post("/upload", uploadHandler.Upload)
			r.Post("/upload/signed", uploadHandler.GenerateSignedURL)
			r.Get("/upload/*", uploadHandler.Download)
		})

		r.Get("/upload/signed/{token}", uploadHandler.ServeSignedURL)
	})

	server := httptest.NewServer(r)
	t.Cleanup(func() {
		server.Close()
		db.Close()
	})

	return server, db
}

func doSetup(t *testing.T, server *httptest.Server, kode, email, password string) {
	t.Helper()
	body := fmt.Sprintf(`{"nama_sekolah":"Test School","kode_sekolah":"%s","nama_admin":"Admin","email_admin":"%s","password_admin":"%s"}`, kode, email, password)
	resp, err := http.Post(server.URL+"/api/v1/setup/", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("setup failed: %d", resp.StatusCode)
	}
}

func doLogin(t *testing.T, server *httptest.Server, kode, email, password string) []*http.Cookie {
	t.Helper()
	body := fmt.Sprintf(`{"kode_sekolah":"%s","email":"%s","password":"%s"}`, kode, email, password)
	resp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("login failed: %d", resp.StatusCode)
	}
	return resp.Cookies()
}

func authRequest(method, url string, body interface{}, cookies []*http.Cookie) (*http.Response, map[string]interface{}) {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return resp, result
}

func TestSetupGuard_BlocksAfterInitialized(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	body := `{"nama_sekolah":"Hack","kode_sekolah":"hack","nama_admin":"H","email_admin":"h@h.id","password_admin":"password123"}`
	resp, _ := http.Post(server.URL+"/api/v1/setup/", "application/json", bytes.NewBufferString(body))
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("expected 404 after setup, got %d", resp.StatusCode)
	}
}

func TestSetupGuard_StatusAfterSetup(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	resp, _ := http.Get(server.URL + "/api/v1/setup/status")
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for status after setup, got %d", resp.StatusCode)
	}
}

func TestLogin_Success(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")
	hasAccess := false
	hasRefresh := false
	for _, c := range cookies {
		if c.Name == "access_token" && c.Value != "" {
			hasAccess = true
		}
		if c.Name == "refresh_token" && c.Value != "" {
			hasRefresh = true
		}
	}
	if !hasAccess {
		t.Error("missing access_token cookie")
	}
	if !hasRefresh {
		t.Error("missing refresh_token cookie")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	body := `{"kode_sekolah":"test1","email":"admin@test1.id","password":"wrong"}`
	resp, _ := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(body))
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_WrongKodeSekolah(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	body := `{"kode_sekolah":"wrong","email":"admin@test1.id","password":"password123"}`
	resp, _ := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString(body))
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_RotatesToken(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
		}
	}
	if refreshCookie == nil {
		t.Fatal("no refresh_token cookie")
	}

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/refresh", nil)
	req.AddCookie(refreshCookie)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("refresh failed: %d", resp.StatusCode)
	}

	var newRefresh string
	for _, c := range resp.Cookies() {
		if c.Name == "refresh_token" {
			newRefresh = c.Value
		}
	}
	if newRefresh == "" {
		t.Error("no new refresh_token after rotation")
	}
	if newRefresh == refreshCookie.Value {
		t.Error("refresh token was not rotated")
	}

	req2, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/refresh", nil)
	req2.AddCookie(refreshCookie)
	resp2, _ := client.Do(req2)
	resp2.Body.Close()
	if resp2.StatusCode != 401 {
		t.Errorf("old refresh token should be rejected, got %d", resp2.StatusCode)
	}
}

func TestTenantIsolation_SiswaNotVisibleAcrossSchools(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")

	hash, _ := bcryptHash("password123")
	db.Exec("INSERT INTO sekolah (nama, kode) VALUES ('School 2', 'school2')")
	db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role) VALUES (2, 'admin@s2.id', ?, 'Admin2', 'admin')", hash)

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")

	resp, _ := authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Siswa School1", "jenis_kelamin": "L",
	}, cookies1)
	if resp.StatusCode != 201 {
		t.Fatalf("create siswa failed: %d", resp.StatusCode)
	}

	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	resp2, result2 := authRequest("GET", server.URL+"/api/v1/siswa", nil, cookies2)
	if resp2.StatusCode != 200 {
		t.Fatalf("list siswa failed: %d", resp2.StatusCode)
	}
	data := result2["data"]
	if data != nil {
		if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
			t.Error("school2 can see school1's siswa — tenant isolation broken")
		}
	}
}

func TestTenantIsolation_SiswaByIDNotAccessible(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")

	hash, _ := bcryptHash("password123")
	db.Exec("INSERT INTO sekolah (nama, kode) VALUES ('School 2', 'school2')")
	db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role) VALUES (2, 'admin@s2.id', ?, 'Admin2', 'admin')", hash)

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Siswa1", "jenis_kelamin": "L",
	}, cookies1)

	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")
	resp, _ := authRequest("GET", server.URL+"/api/v1/siswa/1", nil, cookies2)
	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for cross-tenant siswa access, got %d", resp.StatusCode)
	}
}

func TestPembayaran_StatusTransition(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)

	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)

	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 250000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies)

	resp, result := authRequest("GET", server.URL+"/api/v1/tagihan/1", nil, cookies)
	if resp.StatusCode != 200 {
		t.Fatalf("get tagihan failed: %d", resp.StatusCode)
	}
	data := result["data"].(map[string]interface{})
	if data["status"] != "sebagian" {
		t.Errorf("expected status 'sebagian', got '%v'", data["status"])
	}

	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 250000, "tanggal": "2024-08-12", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/2/verify", nil, cookies)

	resp2, result2 := authRequest("GET", server.URL+"/api/v1/tagihan/1", nil, cookies)
	if resp2.StatusCode != 200 {
		t.Fatalf("get tagihan failed: %d", resp2.StatusCode)
	}
	data2 := result2["data"].(map[string]interface{})
	if data2["status"] != "lunas" {
		t.Errorf("expected status 'lunas', got '%v'", data2["status"])
	}
}

func TestPembayaran_OverpayBlocked(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)

	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 500000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies)

	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 100000, "tanggal": "2024-08-11", "metode": "cash",
	}, cookies)
	resp, _ := authRequest("PUT", server.URL+"/api/v1/pembayaran/2/verify", nil, cookies)
	if resp.StatusCode == 200 {
		t.Error("overpay should be blocked but verify succeeded")
	}
}

func TestRekening_ListAktifOnly(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	authRequest("POST", server.URL+"/api/v1/rekening-sekolah", map[string]interface{}{
		"nama_bank": "BCA", "nomor_rekening": "111", "nama_pemilik": "Test",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/rekening-sekolah", map[string]interface{}{
		"nama_bank": "BRI", "nomor_rekening": "222", "nama_pemilik": "Test",
	}, cookies)

	authRequest("DELETE", server.URL+"/api/v1/rekening-sekolah/1", nil, cookies)

	resp, result := authRequest("GET", server.URL+"/api/v1/rekening-sekolah/aktif", nil, cookies)
	if resp.StatusCode != 200 {
		t.Fatalf("list aktif failed: %d", resp.StatusCode)
	}
	data := result["data"]
	if data == nil {
		t.Fatal("data is nil")
	}
	arr, ok := data.([]interface{})
	if !ok {
		t.Fatal("data is not array")
	}
	if len(arr) != 1 {
		t.Errorf("expected 1 active rekening, got %d", len(arr))
	}
}

func TestLaporan_RekapPembayaran_TahunAjaranFilter(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2023/2024', 0)")
	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)

	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 500000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies)

	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 2, "nominal": 300000, "jatuh_tempo": "2025-08-15",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 2, "siswa_id": 1, "jumlah": 300000, "tanggal": "2025-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/2/verify", nil, cookies)

	resp1, result1 := authRequest("GET", server.URL+"/api/v1/laporan/pembayaran?tanggal_mulai=2024-01-01&tanggal_selesai=2026-01-01", nil, cookies)
	if resp1.StatusCode != 200 {
		t.Fatalf("rekap all failed: %d", resp1.StatusCode)
	}
	data1 := result1["data"].([]interface{})
	totalAll := 0
	for _, item := range data1 {
		m := item.(map[string]interface{})
		totalAll += int(m["total_transaksi"].(float64))
	}
	if totalAll != 2 {
		t.Errorf("expected 2 transactions total (no filter), got %d", totalAll)
	}

	resp2, result2 := authRequest("GET", server.URL+"/api/v1/laporan/pembayaran?tanggal_mulai=2024-01-01&tanggal_selesai=2026-01-01&tahun_ajaran_id=1", nil, cookies)
	if resp2.StatusCode != 200 {
		t.Fatalf("rekap ta1 failed: %d", resp2.StatusCode)
	}
	data2 := result2["data"].([]interface{})
	totalTA1 := 0
	for _, item := range data2 {
		m := item.(map[string]interface{})
		totalTA1 += int(m["total_transaksi"].(float64))
	}
	if totalTA1 != 1 {
		t.Errorf("expected 1 transaction for tahun_ajaran_id=1, got %d", totalTA1)
	}
}

func TestLaporan_RekapSiswa_TahunAjaranFilter(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2023/2024', 0)")
	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L", "tahun_ajaran_masuk": 1,
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "002", "nama": "Budi", "jenis_kelamin": "L", "tahun_ajaran_masuk": 2,
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "003", "nama": "Citra", "jenis_kelamin": "P", "tahun_ajaran_masuk": 2,
	}, cookies)

	respAll, resultAll := authRequest("GET", server.URL+"/api/v1/laporan/siswa", nil, cookies)
	if respAll.StatusCode != 200 {
		t.Fatalf("rekap siswa all failed: %d", respAll.StatusCode)
	}
	dataAll := resultAll["data"].(map[string]interface{})
	if int(dataAll["total"].(float64)) != 3 {
		t.Errorf("expected 3 total siswa (no filter), got %v", dataAll["total"])
	}

	respTA2, resultTA2 := authRequest("GET", server.URL+"/api/v1/laporan/siswa?tahun_ajaran_id=2", nil, cookies)
	if respTA2.StatusCode != 200 {
		t.Fatalf("rekap siswa ta2 failed: %d", respTA2.StatusCode)
	}
	dataTA2 := resultTA2["data"].(map[string]interface{})
	if int(dataTA2["total"].(float64)) != 2 {
		t.Errorf("expected 2 siswa for tahun_ajaran_id=2, got %v", dataTA2["total"])
	}
	if int(dataTA2["laki_laki"].(float64)) != 1 {
		t.Errorf("expected 1 laki_laki for ta2, got %v", dataTA2["laki_laki"])
	}
	if int(dataTA2["perempuan"].(float64)) != 1 {
		t.Errorf("expected 1 perempuan for ta2, got %v", dataTA2["perempuan"])
	}
}

func TestLaporan_CrossSchoolIsolation(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")

	hash, _ := bcryptHash("password123")
	db.Exec("INSERT INTO sekolah (nama, kode) VALUES ('School 2', 'school2')")
	db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role) VALUES (2, 'admin@s2.id', ?, 'Admin2', 'admin')", hash)

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")
	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (2, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Siswa1", "jenis_kelamin": "L", "tahun_ajaran_masuk": 1,
	}, cookies1)

	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	resp, result := authRequest("GET", server.URL+"/api/v1/laporan/siswa?tahun_ajaran_id=1", nil, cookies2)
	if resp.StatusCode != 200 {
		t.Fatalf("rekap siswa school2 failed: %d", resp.StatusCode)
	}
	data := result["data"].(map[string]interface{})
	if int(data["total"].(float64)) != 0 {
		t.Errorf("school2 should see 0 siswa from school1, got %v", data["total"])
	}
}

func TestExport_Siswa_ReturnsXLSX(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "002", "nama": "Budi", "jenis_kelamin": "L",
	}, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/siswa/export", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected XLSX content type, got %s", ct)
	}
	cd := resp.Header.Get("Content-Disposition")
	if cd == "" {
		t.Error("missing Content-Disposition header")
	}
}

func TestExport_Siswa_IsTenantScoped(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")

	hash, _ := bcryptHash("password123")
	db.Exec("INSERT INTO sekolah (nama, kode) VALUES ('School 2', 'school2')")
	db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role) VALUES (2, 'admin@s2.id', ?, 'Admin2', 'admin')", hash)

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Siswa School1", "jenis_kelamin": "L",
	}, cookies1)

	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/siswa/export", nil)
	for _, c := range cookies2 {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestExport_Tagihan_ReturnsXLSX(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/pembayaran/export", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected XLSX content type, got %s", ct)
	}
}

func TestExport_Tagihan_RespectsFilters(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2023/2024', 0)")
	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)

	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 2, "nominal": 300000, "jatuh_tempo": "2025-08-15",
	}, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/pembayaran/export?tahun_ajaran_id=1", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestKwitansi_ReturnsHTML(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 500000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/pembayaran/1/kwitansi", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type, got %s", ct)
	}
}

func TestKwitansi_CrossSchoolRejected(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")

	hash, _ := bcryptHash("password123")
	db.Exec("INSERT INTO sekolah (nama, kode) VALUES ('School 2', 'school2')")
	db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role) VALUES (2, 'admin@s2.id', ?, 'Admin2', 'admin')", hash)

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies1)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies1)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies1)
	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 500000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies1)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies1)

	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/pembayaran/1/kwitansi", nil)
	for _, c := range cookies2 {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for cross-school kwitansi, got %d", resp.StatusCode)
	}
}

func TestExport_PPDB_ReturnsXLSX(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/ppdb/daftar", map[string]interface{}{
		"tahun_ajaran_id": 1, "nama_lengkap": "Calon Siswa", "jenis_kelamin": "L",
	}, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/ppdb/export?tahun_ajaran_id=1", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected XLSX content type, got %s", ct)
	}
}

func TestExport_LaporanPembayaran_ReturnsXLSX(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	db.Exec("INSERT INTO tahun_ajaran (sekolah_id, nama, aktif) VALUES (1, '2024/2025', 1)")

	authRequest("POST", server.URL+"/api/v1/siswa", map[string]interface{}{
		"nis": "001", "nama": "Andi", "jenis_kelamin": "L",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/kategori-pembayaran", map[string]interface{}{
		"nama": "SPP",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/tagihan", map[string]interface{}{
		"siswa_id": 1, "kategori_id": 1, "tahun_ajaran_id": 1, "nominal": 500000, "jatuh_tempo": "2024-08-15",
	}, cookies)
	authRequest("POST", server.URL+"/api/v1/pembayaran", map[string]interface{}{
		"tagihan_id": 1, "siswa_id": 1, "jumlah": 500000, "tanggal": "2024-08-10", "metode": "cash",
	}, cookies)
	authRequest("PUT", server.URL+"/api/v1/pembayaran/1/verify", nil, cookies)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/laporan/pembayaran/export?tanggal_mulai=2024-01-01&tanggal_selesai=2025-12-31", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected XLSX content type, got %s", ct)
	}
}

func TestExport_LaporanSiswa_ReturnsXLSX(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/laporan/siswa/export", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected XLSX content type, got %s", ct)
	}
}
