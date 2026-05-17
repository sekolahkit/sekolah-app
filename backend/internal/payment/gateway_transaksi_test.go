package payment

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func setupTestDBWithInitiation(t *testing.T) *sql.DB {
	t.Helper()
	db := setupTestDB(t)

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS pengguna (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			email TEXT NOT NULL,
			nama TEXT NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS siswa (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nis TEXT NOT NULL,
			nama TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			aktif INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestInitiateTransactionProviderNotConfigured(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	service := NewService(repo)

	_, err = service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != ErrProviderNotConfig {
		t.Fatalf("expected ErrProviderNotConfig, got: %v", err)
	}
}

func TestInitiateTransactionTagihanLunas(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO siswa (id, sekolah_id, nis, nama) VALUES (1, 1, '001', 'Siswa 1')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO kategori_pembayaran (id, sekolah_id, nama) VALUES (1, 1, 'SPP')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, kategori_id, tahun_ajaran_id, nominal, status) VALUES (1, 1, 1, 1, 1, 100000, 'lunas')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != ErrTagihanLunas {
		t.Fatalf("expected ErrTagihanLunas, got: %v", err)
	}
}

func TestInitiateTransactionSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO siswa (id, sekolah_id, nis, nama) VALUES (1, 1, '001', 'Siswa 1')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO kategori_pembayaran (id, sekolah_id, nama) VALUES (1, 1, 'SPP')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, kategori_id, tahun_ajaran_id, nominal, status) VALUES (1, 1, 1, 1, 1, 100000, 'belum_bayar')`)
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Provider != "midtrans" {
		t.Errorf("expected provider midtrans, got: %s", result.Provider)
	}
	if result.OrderID != "1" {
		t.Errorf("expected order_id 1, got: %s", result.OrderID)
	}
	if result.PaymentURL == "" {
		t.Error("expected non-empty payment_url")
	}
	if result.Status != "pending" {
		t.Errorf("expected status pending, got: %s", result.Status)
	}
	if result.Amount != 100000 {
		t.Errorf("expected amount 100000, got: %d", result.Amount)
	}
}

func TestInitiateTransactionIdempotency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO siswa (id, sekolah_id, nis, nama) VALUES (1, 1, '001', 'Siswa 1')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO kategori_pembayaran (id, sekolah_id, nama) VALUES (1, 1, 'SPP')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, kategori_id, tahun_ajaran_id, nominal, status) VALUES (1, 1, 1, 1, 1, 100000, 'belum_bayar')`)
	if err != nil {
		t.Fatal(err)
	}

	result1, err := service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != nil {
		t.Fatalf("first initiation failed: %v", err)
	}

	result2, err := service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != nil {
		t.Fatalf("second initiation failed: %v", err)
	}

	if result1.ID != result2.ID {
		t.Errorf("expected same transaction ID for idempotency, got %d and %d", result1.ID, result2.ID)
	}
	if result1.OrderID != result2.OrderID {
		t.Errorf("expected same order_id for idempotency, got %s and %s", result1.OrderID, result2.OrderID)
	}
}

func TestInitiateTransactionTagihanNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.InitiateTransaction(1, 999, 1, "midtrans")
	if err == nil {
		t.Fatal("expected error for non-existent tagihan")
	}
}

func TestInitiateTransactionCrossSchool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test'), (2, 'Other')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO siswa (id, sekolah_id, nis, nama) VALUES (1, 1, '001', 'Siswa 1')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO kategori_pembayaran (id, sekolah_id, nama) VALUES (1, 1, 'SPP')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, kategori_id, tahun_ajaran_id, nominal, status) VALUES (1, 1, 1, 1, 1, 100000, 'belum_bayar')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.InitiateTransaction(2, 1, 1, "midtrans")
	if err == nil {
		t.Fatal("expected error for cross-school access")
	}
}

func TestInitiateTransactionInvalidProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	service := NewService(repo, gw)

	_, err = service.InitiateTransaction(1, 1, 1, "invalid")
	if err != ErrProviderNotConfig {
		t.Fatalf("expected ErrProviderNotConfig, got: %v", err)
	}
}

func TestAvailableProviders(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	midtrans := NewMidtrans(MidtransConfig{ServerKey: "test-key"})
	xendit := NewXendit(XenditConfig{SecretKey: "secret", CallbackToken: "token"})
	service := NewService(repo, midtrans, xendit)

	providers := service.AvailableProviders()
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got: %d", len(providers))
	}

	if !service.HasProvider("midtrans") {
		t.Error("expected midtrans to be available")
	}
	if !service.HasProvider("xendit") {
		t.Error("expected xendit to be available")
	}
	if service.HasProvider("unknown") {
		t.Error("expected unknown to not be available")
	}
}

func TestCallbackUpdatesGatewayTransaksi(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nis TEXT NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS kategori_pembayaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE IF NOT EXISTS gateway_transaksi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tagihan_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_gateway_id TEXT,
			payment_url TEXT NOT NULL,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			expires_at DATETIME,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(tagihan_id, provider, status)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	serverKey := "test-key"
	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})
	service := NewService(repo, gw)

	_, err = db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'test@test.com', 'Test', 'hash', 'admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO siswa (id, sekolah_id, nis, nama) VALUES (1, 1, '001', 'Siswa 1')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO kategori_pembayaran (id, sekolah_id, nama) VALUES (1, 1, 'SPP')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, kategori_id, tahun_ajaran_id, nominal, status) VALUES (1, 1, 1, 1, 1, 100000, 'belum_bayar')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.InitiateTransaction(1, 1, 1, "midtrans")
	if err != nil {
		t.Fatalf("initiation failed: %v", err)
	}

	orderID := "1"
	statusCode := "200"
	grossAmount := "100000"
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.New()
	h.Write([]byte(raw))
	sig := hex.EncodeToString(h.Sum(nil))

	notif := midtransNotification{
		TransactionID:     "txn-001",
		OrderID:           orderID,
		TransactionStatus: "settlement",
		FraudStatus:       "accept",
		StatusCode:        statusCode,
		GrossAmount:       grossAmount,
		SignatureKey:      sig,
	}
	payload, _ := json.Marshal(notif)

	result, err := service.ProcessCallback("midtrans", payload, nil)
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Errorf("expected StatusSuccess, got: %s", result.Status)
	}

	gt, err := repo.GetGatewayTransaksiByOrderID(orderID)
	if err != nil {
		t.Fatalf("get gateway transaksi failed: %v", err)
	}
	if gt == nil {
		t.Fatal("expected gateway_transaksi to exist")
	}
	if gt.Status != "paid" {
		t.Errorf("expected gateway_transaksi status paid, got: %s", gt.Status)
	}
}
