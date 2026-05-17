package notifikasi

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Notifikasi struct {
	ID          int64  `json:"id"`
	SekolahID   int64  `json:"sekolah_id"`
	Tipe        string `json:"tipe"`
	Penerima    string `json:"penerima"`
	Pesan       string `json:"pesan"`
	Status      string `json:"status"`
	RetryCount  int    `json:"retry_count"`
	MaxRetries  int    `json:"max_retries"`
	LastError   string `json:"last_error"`
	ScheduledAt string `json:"scheduled_at"`
	SentAt      string `json:"sent_at"`
	CreatedAt   string `json:"created_at"`
}

type QueueStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Sent       int `json:"sent"`
	Failed     int `json:"failed"`
}

type ListParams struct {
	Page   int
	Limit  int
	Status string
	Tipe   string
}

func (r *Repository) List(sekolahID int64, params ListParams) ([]Notifikasi, int, error) {
	query := sq.Select("id", "sekolah_id", "tipe", "penerima", "pesan", "status",
		"retry_count", "max_retries", "COALESCE(last_error,'')",
		"COALESCE(scheduled_at,'')", "COALESCE(sent_at,'')", "created_at").
		From("notifikasi_antrian").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("notifikasi_antrian").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Status != "" {
		query = query.Where(sq.Eq{"status": params.Status})
		countQuery = countQuery.Where(sq.Eq{"status": params.Status})
	}
	if params.Tipe != "" {
		query = query.Where(sq.Eq{"tipe": params.Tipe})
		countQuery = countQuery.Where(sq.Eq{"tipe": params.Tipe})
	}

	var total int
	err := countQuery.RunWith(r.db).QueryRow().Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query = query.OrderBy("created_at DESC").Limit(uint64(params.Limit)).Offset(uint64(offset))

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Notifikasi
	for rows.Next() {
		var n Notifikasi
		err := rows.Scan(&n.ID, &n.SekolahID, &n.Tipe, &n.Penerima, &n.Pesan, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.LastError, &n.ScheduledAt, &n.SentAt, &n.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, n)
	}
	return list, total, nil
}

func (r *Repository) Create(n *Notifikasi) (int64, error) {
	result, err := sq.Insert("notifikasi_antrian").
		Columns("sekolah_id", "tipe", "penerima", "pesan", "status", "max_retries", "scheduled_at").
		Values(n.SekolahID, n.Tipe, n.Penerima, n.Pesan, n.Status, n.MaxRetries, n.ScheduledAt).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateStatus(id int64, status, lastError string, retryCount int) error {
	builder := sq.Update("notifikasi_antrian").
		Set("status", status).
		Set("last_error", lastError).
		Set("retry_count", retryCount).
		Where(sq.Eq{"id": id})

	if status == "sent" {
		builder = builder.Set("sent_at", sq.Expr("CURRENT_TIMESTAMP"))
	}

	if status != "processing" {
		builder = builder.Set("claimed_at", nil)
	}

	_, err := builder.RunWith(r.db).Exec()
	return err
}

func (r *Repository) GetByID(sekolahID, id int64) (*Notifikasi, error) {
	var n Notifikasi
	err := sq.Select("id", "sekolah_id", "tipe", "penerima", "pesan", "status",
		"retry_count", "max_retries", "COALESCE(last_error,'')",
		"COALESCE(scheduled_at,'')", "COALESCE(sent_at,'')", "created_at").
		From("notifikasi_antrian").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&n.ID, &n.SekolahID, &n.Tipe, &n.Penerima, &n.Pesan, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.LastError,
			&n.ScheduledAt, &n.SentAt, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *Repository) ResetForRetry(id int64) error {
	_, err := sq.Update("notifikasi_antrian").
		Set("status", "pending").
		Set("last_error", "").
		Set("claimed_at", nil).
		Where(sq.Eq{"id": id}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ClaimPending(limit int) ([]Notifikasi, error) {
	claimToken := time.Now().UTC().Format("2006-01-02 15:04:05.000000000")

	res, err := sq.Update("notifikasi_antrian").
		Set("status", "processing").
		Set("claimed_at", claimToken).
		Where("id IN (SELECT id FROM notifikasi_antrian WHERE status = 'pending' AND retry_count < max_retries AND (scheduled_at IS NULL OR scheduled_at <= datetime('now')) ORDER BY created_at ASC LIMIT ?)", limit).
		RunWith(r.db).Exec()
	if err != nil {
		return nil, err
	}

	claimed, _ := res.RowsAffected()
	if claimed == 0 {
		return nil, nil
	}

	rows, err := sq.Select("id", "sekolah_id", "tipe", "penerima", "pesan", "status",
		"retry_count", "max_retries", "COALESCE(last_error,'')",
		"COALESCE(scheduled_at,'')", "COALESCE(sent_at,'')", "created_at").
		From("notifikasi_antrian").
		Where(sq.Eq{"status": "processing", "claimed_at": claimToken}).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Notifikasi
	for rows.Next() {
		var n Notifikasi
		err := rows.Scan(&n.ID, &n.SekolahID, &n.Tipe, &n.Penerima, &n.Pesan, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.LastError, &n.ScheduledAt, &n.SentAt, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *Repository) ReleaseStale(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan).UTC().Format("2006-01-02 15:04:05")
	res, err := sq.Update("notifikasi_antrian").
		Set("status", "pending").
		Set("claimed_at", nil).
		Where(sq.Eq{"status": "processing"}).
		Where("claimed_at < ?", cutoff).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) GetQueueStats(sekolahID int64) (*QueueStats, error) {
	rows, err := sq.Select("status", "COUNT(*)").
		From("notifikasi_antrian").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		GroupBy("status").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &QueueStats{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		switch status {
		case "pending":
			stats.Pending = count
		case "processing":
			stats.Processing = count
		case "sent":
			stats.Sent = count
		case "failed":
			stats.Failed = count
		}
	}
	return stats, nil
}

func (r *Repository) GetPending(sekolahID int64, limit int) ([]Notifikasi, error) {
	rows, err := sq.Select("id", "sekolah_id", "tipe", "penerima", "pesan", "status",
		"retry_count", "max_retries", "COALESCE(last_error,'')",
		"COALESCE(scheduled_at,'')", "COALESCE(sent_at,'')", "created_at").
		From("notifikasi_antrian").
		Where(sq.Eq{"sekolah_id": sekolahID, "status": "pending"}).
		Where("retry_count < max_retries").
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Notifikasi
	for rows.Next() {
		var n Notifikasi
		err := rows.Scan(&n.ID, &n.SekolahID, &n.Tipe, &n.Penerima, &n.Pesan, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.LastError, &n.ScheduledAt, &n.SentAt, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *Repository) GetAllPending(limit int) ([]Notifikasi, error) {
	rows, err := sq.Select("id", "sekolah_id", "tipe", "penerima", "pesan", "status",
		"retry_count", "max_retries", "COALESCE(last_error,'')",
		"COALESCE(scheduled_at,'')", "COALESCE(sent_at,'')", "created_at").
		From("notifikasi_antrian").
		Where(sq.Eq{"status": "pending"}).
		Where("retry_count < max_retries").
		Where("scheduled_at IS NULL OR scheduled_at <= datetime('now')").
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Notifikasi
	for rows.Next() {
		var n Notifikasi
		err := rows.Scan(&n.ID, &n.SekolahID, &n.Tipe, &n.Penerima, &n.Pesan, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.LastError, &n.ScheduledAt, &n.SentAt, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}
