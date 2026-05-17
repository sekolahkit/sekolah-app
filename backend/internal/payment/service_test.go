package payment

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tagihan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			siswa_id INTEGER NOT NULL,
			kategori_id INTEGER NOT NULL DEFAULT 0,
			tahun_ajaran_id INTEGER NOT NULL DEFAULT 0,
			semester TEXT DEFAULT '',
			nominal INTEGER NOT NULL,
			jatuh_tempo TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'belum_bayar',
			catatan TEXT DEFAULT '',
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS pembayaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tagihan_id INTEGER NOT NULL,
			siswa_id INTEGER NOT NULL,
			sekolah_id INTEGER NOT NULL,
			jumlah INTEGER NOT NULL,
			tanggal TEXT DEFAULT '',
			metode TEXT DEFAULT '',
			provider TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			bukti_bayar TEXT DEFAULT '',
			rekening_sekolah_id INTEGER DEFAULT 0,
			catatan TEXT DEFAULT '',
			verified_by INTEGER DEFAULT 0,
			verified_at TEXT DEFAULT '',
			created_at DATETIME DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS payment_callback_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider TEXT NOT NULL,
			payment_gateway_id TEXT NOT NULL,
			order_id TEXT NOT NULL,
			status TEXT NOT NULL,
			amount INTEGER NOT NULL DEFAULT 0,
			sekolah_id INTEGER NOT NULL,
			processed INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT (datetime('now')),
			UNIQUE(provider, payment_gateway_id)
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

func insertTestTagihan(t *testing.T, db *sql.DB, id, sekolahID, siswaID, nominal int64) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO tagihan (id, sekolah_id, siswa_id, nominal, status) VALUES (?, ?, ?, ?, 'belum_bayar')`,
		id, sekolahID, siswaID, nominal)
	if err != nil {
		t.Fatal(err)
	}
}

func midtransPayload(t *testing.T, serverKey, orderID, statusCode, grossAmount, txStatus string) []byte {
	t.Helper()
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.New()
	h.Write([]byte(raw))
	sig := hex.EncodeToString(h.Sum(nil))

	notif := midtransNotification{
		TransactionID:     "txn-" + orderID,
		OrderID:           orderID,
		TransactionStatus: txStatus,
		FraudStatus:       "accept",
		StatusCode:        statusCode,
		GrossAmount:       grossAmount,
		SignatureKey:       sig,
	}
	payload, _ := json.Marshal(notif)
	return payload
}

func TestServiceDuplicateCallbackIdempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestTagihan(t, db, 1, 1, 10, 100000)

	serverKey := "test-key"
	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})
	svc := NewService(repo, gw)

	payload := midtransPayload(t, serverKey, "1", "200", "50000", "settlement")

	result, err := svc.ProcessCallback("midtrans", payload, nil)
	if err != nil {
		t.Fatalf("first callback failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected success, got %s", result.Status)
	}

	_, err = svc.ProcessCallback("midtrans", payload, nil)
	if err != ErrDuplicateCallback {
		t.Fatalf("expected ErrDuplicateCallback, got: %v", err)
	}
}

func TestServiceOverpayRejected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestTagihan(t, db, 1, 1, 10, 50000)

	serverKey := "test-key"
	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})
	svc := NewService(repo, gw)

	payload := midtransPayload(t, serverKey, "1", "200", "60000", "settlement")

	_, err := svc.ProcessCallback("midtrans", payload, nil)
	if err != ErrOverpay {
		t.Fatalf("expected ErrOverpay, got: %v", err)
	}
}

func TestServiceValidCallbackUpdatesPembayaran(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestTagihan(t, db, 1, 1, 10, 100000)

	serverKey := "test-key"
	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})
	svc := NewService(repo, gw)

	payload := midtransPayload(t, serverKey, "1", "200", "100000", "settlement")

	result, err := svc.ProcessCallback("midtrans", payload, nil)
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected success, got %s", result.Status)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM pembayaran WHERE tagihan_id = 1 AND status = 'verified'`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 pembayaran, got %d", count)
	}

	var tagihanStatus string
	err = db.QueryRow(`SELECT status FROM tagihan WHERE id = 1`).Scan(&tagihanStatus)
	if err != nil {
		t.Fatal(err)
	}
	if tagihanStatus != "lunas" {
		t.Fatalf("expected tagihan status lunas, got %s", tagihanStatus)
	}
}

func TestServicePartialPaymentUpdatesStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestTagihan(t, db, 1, 1, 10, 100000)

	serverKey := "test-key"
	repo := NewRepository(db)
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})
	svc := NewService(repo, gw)

	payload := midtransPayload(t, serverKey, "1", "200", "50000", "settlement")

	_, err := svc.ProcessCallback("midtrans", payload, nil)
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}

	var tagihanStatus string
	err = db.QueryRow(`SELECT status FROM tagihan WHERE id = 1`).Scan(&tagihanStatus)
	if err != nil {
		t.Fatal(err)
	}
	if tagihanStatus != "sebagian" {
		t.Fatalf("expected tagihan status sebagian, got %s", tagihanStatus)
	}
}

func TestServiceXenditValidCallback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestTagihan(t, db, 1, 1, 10, 75000)

	token := "xnd-token"
	repo := NewRepository(db)
	gw := NewXendit(XenditConfig{SecretKey: "s", CallbackToken: token})
	svc := NewService(repo, gw)

	notif := xenditNotification{
		ID:         "xnd-pay-001",
		ExternalID: "1",
		Status:     "PAID",
		Amount:     75000,
	}
	payload, _ := json.Marshal(notif)
	headers := map[string]string{"X-Callback-Token": token}

	result, err := svc.ProcessCallback("xendit", payload, headers)
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected success, got %s", result.Status)
	}

	var tagihanStatus string
	err = db.QueryRow(`SELECT status FROM tagihan WHERE id = 1`).Scan(&tagihanStatus)
	if err != nil {
		t.Fatal(err)
	}
	if tagihanStatus != "lunas" {
		t.Fatalf("expected lunas, got %s", tagihanStatus)
	}
}
