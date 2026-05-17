package tahun_ajaran

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type TahunAjaran struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	Nama          string `json:"nama"`
	Aktif         bool   `json:"aktif"`
	TanggalMulai  string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
	CreatedAt     string `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(sekolahID int64) ([]TahunAjaran, error) {
	rows, err := sq.Select("id", "sekolah_id", "nama", "aktif",
		"COALESCE(tanggal_mulai,'')", "COALESCE(tanggal_selesai,'')", "created_at").
		From("tahun_ajaran").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		OrderBy("aktif DESC, nama DESC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []TahunAjaran
	for rows.Next() {
		var t TahunAjaran
		if err := rows.Scan(&t.ID, &t.SekolahID, &t.Nama, &t.Aktif,
			&t.TanggalMulai, &t.TanggalSelesai, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *Repository) GetAktif(sekolahID int64) (*TahunAjaran, error) {
	var t TahunAjaran
	err := sq.Select("id", "sekolah_id", "nama", "aktif",
		"COALESCE(tanggal_mulai,'')", "COALESCE(tanggal_selesai,'')", "created_at").
		From("tahun_ajaran").
		Where(sq.Eq{"sekolah_id": sekolahID, "aktif": true}).
		RunWith(r.db).QueryRow().
		Scan(&t.ID, &t.SekolahID, &t.Nama, &t.Aktif,
			&t.TanggalMulai, &t.TanggalSelesai, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*TahunAjaran, error) {
	var t TahunAjaran
	err := sq.Select("id", "sekolah_id", "nama", "aktif",
		"COALESCE(tanggal_mulai,'')", "COALESCE(tanggal_selesai,'')", "created_at").
		From("tahun_ajaran").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&t.ID, &t.SekolahID, &t.Nama, &t.Aktif,
			&t.TanggalMulai, &t.TanggalSelesai, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Create(t *TahunAjaran) (int64, error) {
	result, err := sq.Insert("tahun_ajaran").
		Columns("sekolah_id", "nama", "aktif", "tanggal_mulai", "tanggal_selesai").
		Values(t.SekolahID, t.Nama, t.Aktif, nullStr(t.TanggalMulai), nullStr(t.TanggalSelesai)).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(t *TahunAjaran) error {
	_, err := sq.Update("tahun_ajaran").
		Set("nama", t.Nama).
		Set("tanggal_mulai", nullStr(t.TanggalMulai)).
		Set("tanggal_selesai", nullStr(t.TanggalSelesai)).
		Where(sq.Eq{"id": t.ID, "sekolah_id": t.SekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) SetAktif(sekolahID, id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := sq.Update("tahun_ajaran").
		Set("aktif", false).
		Where(sq.Eq{"sekolah_id": sekolahID}).
		RunWith(tx).Exec(); err != nil {
		return err
	}

	if _, err := sq.Update("tahun_ajaran").
		Set("aktif", true).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(tx).Exec(); err != nil {
		return err
	}

	return tx.Commit()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	Nama           string `json:"nama"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
}

type UpdateRequest struct {
	Nama           string `json:"nama"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
}

func (s *Service) List(sekolahID int64) ([]TahunAjaran, error) {
	return s.repo.List(sekolahID)
}

func (s *Service) GetAktif(sekolahID int64) (*TahunAjaran, error) {
	return s.repo.GetAktif(sekolahID)
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*TahunAjaran, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	t := &TahunAjaran{
		SekolahID:      sekolahID,
		Nama:           req.Nama,
		TanggalMulai:   req.TanggalMulai,
		TanggalSelesai: req.TanggalSelesai,
	}

	id, err := s.repo.Create(t)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat tahun ajaran: %w", err)
	}
	t.ID = id
	return t, nil
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*TahunAjaran, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	existing, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		return nil, fmt.Errorf("tahun ajaran tidak ditemukan")
	}

	existing.Nama = req.Nama
	existing.TanggalMulai = req.TanggalMulai
	existing.TanggalSelesai = req.TanggalSelesai

	if err := s.repo.Update(existing); err != nil {
		return nil, fmt.Errorf("gagal mengupdate tahun ajaran: %w", err)
	}
	return existing, nil
}

func (s *Service) SetAktif(sekolahID, id int64) error {
	if _, err := s.repo.GetByID(sekolahID, id); err != nil {
		return fmt.Errorf("tahun ajaran tidak ditemukan")
	}
	return s.repo.SetAktif(sekolahID, id)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	list, err := h.service.List(user.SekolahID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data tahun ajaran")
		return
	}
	response.JSON(w, 200, list)
}

func (h *Handler) GetAktif(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	t, err := h.service.GetAktif(user.SekolahID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.JSON(w, 200, nil)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil tahun ajaran aktif")
		return
	}
	response.JSON(w, 200, t)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	var req CreateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}
	t, err := h.service.Create(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}
	response.JSON(w, 201, t)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}
	var req UpdateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}
	t, err := h.service.Update(user.SekolahID, id, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, 200, t)
}

func (h *Handler) SetAktif(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}
	if err := h.service.SetAktif(user.SekolahID, id); err != nil {
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, 200, map[string]string{"message": "Tahun ajaran aktif berhasil diubah"})
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/tahun-ajaran", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/aktif", h.GetAktif)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Put("/{id}/aktif", h.SetAktif)
	})
}
