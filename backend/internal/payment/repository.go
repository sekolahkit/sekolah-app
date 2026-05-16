package payment

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CallbackLog struct {
	ID               int64
	Provider         string
	PaymentGatewayID string
	OrderID          string
	Status           string
	Amount           int64
	SekolahID        int64
	Processed        bool
}

func (r *Repository) FindCallbackLog(provider, paymentGatewayID string) (*CallbackLog, error) {
	row := sq.Select("id", "provider", "payment_gateway_id", "order_id", "status", "amount", "sekolah_id", "processed").
		From("payment_callback_logs").
		Where(sq.Eq{"provider": provider, "payment_gateway_id": paymentGatewayID}).
		RunWith(r.db).
		QueryRow()

	var log CallbackLog
	err := row.Scan(&log.ID, &log.Provider, &log.PaymentGatewayID, &log.OrderID, &log.Status, &log.Amount, &log.SekolahID, &log.Processed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *Repository) InsertCallbackLog(log *CallbackLog) error {
	_, err := sq.Insert("payment_callback_logs").
		Columns("provider", "payment_gateway_id", "order_id", "status", "amount", "sekolah_id", "processed").
		Values(log.Provider, log.PaymentGatewayID, log.OrderID, log.Status, log.Amount, log.SekolahID, log.Processed).
		RunWith(r.db).
		Exec()
	return err
}

type TagihanInfo struct {
	ID        int64
	SekolahID int64
	SiswaID   int64
	Nominal   int64
	Status    string
}

func (r *Repository) FindTagihanByOrderID(orderID string) (*TagihanInfo, error) {
	row := sq.Select("id", "sekolah_id", "siswa_id", "nominal", "status").
		From("tagihan").
		Where(sq.Eq{"id": orderID}).
		RunWith(r.db).
		QueryRow()

	var t TagihanInfo
	err := row.Scan(&t.ID, &t.SekolahID, &t.SiswaID, &t.Nominal, &t.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetPaidAmount(tagihanID int64) (int64, error) {
	row := sq.Select("COALESCE(SUM(jumlah), 0)").
		From("pembayaran").
		Where(sq.Eq{"tagihan_id": tagihanID, "status": "verified"}).
		RunWith(r.db).
		QueryRow()

	var total int64
	err := row.Scan(&total)
	return total, err
}

func (r *Repository) InsertPembayaranAndUpdateTagihan(tagihanID, siswaID, sekolahID, amount int64, provider, paymentGatewayID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = sq.Insert("pembayaran").
		Columns("tagihan_id", "siswa_id", "sekolah_id", "jumlah", "tanggal", "metode", "provider", "status", "bukti_bayar", "rekening_sekolah_id", "catatan", "verified_by", "verified_at").
		Values(tagihanID, siswaID, sekolahID, amount, "datetime('now')", provider, provider, "verified", "", 0, fmt.Sprintf("auto via %s callback", provider), 0, "datetime('now')").
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}

	paidRow := sq.Select("COALESCE(SUM(jumlah), 0)").
		From("pembayaran").
		Where(sq.Eq{"tagihan_id": tagihanID, "status": "verified"}).
		RunWith(tx).
		QueryRow()

	var totalPaid int64
	if err := paidRow.Scan(&totalPaid); err != nil {
		return err
	}

	nominalRow := sq.Select("nominal").
		From("tagihan").
		Where(sq.Eq{"id": tagihanID}).
		RunWith(tx).
		QueryRow()

	var nominal int64
	if err := nominalRow.Scan(&nominal); err != nil {
		return err
	}

	newStatus := "sebagian"
	if totalPaid >= nominal {
		newStatus = "lunas"
	}

	_, err = sq.Update("tagihan").
		Set("status", newStatus).
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"id": tagihanID}).
		RunWith(tx).
		Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}
