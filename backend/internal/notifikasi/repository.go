package notifikasi

import (
	"database/sql"

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
	Pending int `json:"pending"`
	Sent    int `json:"sent"`
	Failed  int `json:"failed"`
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

	_, err := builder.RunWith(r.db).Exec()
	return err
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
