package jurusan

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

type Jurusan struct {
	ID        int64  `json:"id"`
	SekolahID int64  `json:"sekolah_id"`
	Nama      string `json:"nama"`
	Kode      string `json:"kode"`
	CreatedAt string `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(sekolahID int64) ([]Jurusan, error) {
	rows, err := sq.Select("id", "sekolah_id", "nama", "COALESCE(kode,'')", "created_at").
		From("jurusan").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		OrderBy("nama ASC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Jurusan
	for rows.Next() {
		var j Jurusan
		if err := rows.Scan(&j.ID, &j.SekolahID, &j.Nama, &j.Kode, &j.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, j)
	}
	return list, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*Jurusan, error) {
	var j Jurusan
	err := sq.Select("id", "sekolah_id", "nama", "COALESCE(kode,'')", "created_at").
		From("jurusan").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&j.ID, &j.SekolahID, &j.Nama, &j.Kode, &j.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *Repository) Create(j *Jurusan) (int64, error) {
	result, err := sq.Insert("jurusan").
		Columns("sekolah_id", "nama", "kode").
		Values(j.SekolahID, j.Nama, nullStr(j.Kode)).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(j *Jurusan) error {
	_, err := sq.Update("jurusan").
		Set("nama", j.Nama).
		Set("kode", nullStr(j.Kode)).
		Where(sq.Eq{"id": j.ID, "sekolah_id": j.SekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) Delete(sekolahID, id int64) error {
	_, err := sq.Delete("jurusan").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) IsReferenced(id int64) (bool, error) {
	var count int
	err := sq.Select("COUNT(*)").
		From("kelas").
		Where(sq.Eq{"jurusan_id": id}).
		RunWith(r.db).QueryRow().Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}

type UpdateRequest struct {
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}

func (s *Service) List(sekolahID int64) ([]Jurusan, error) {
	return s.repo.List(sekolahID)
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*Jurusan, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	j := &Jurusan{
		SekolahID: sekolahID,
		Nama:      req.Nama,
		Kode:      req.Kode,
	}

	id, err := s.repo.Create(j)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat jurusan: %w", err)
	}
	j.ID = id
	return j, nil
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*Jurusan, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	existing, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		return nil, fmt.Errorf("jurusan tidak ditemukan")
	}

	existing.Nama = req.Nama
	existing.Kode = req.Kode

	if err := s.repo.Update(existing); err != nil {
		return nil, fmt.Errorf("gagal mengupdate jurusan: %w", err)
	}
	return existing, nil
}

func (s *Service) Delete(sekolahID, id int64) error {
	if _, err := s.repo.GetByID(sekolahID, id); err != nil {
		return fmt.Errorf("jurusan tidak ditemukan")
	}

	referenced, err := s.repo.IsReferenced(id)
	if err != nil {
		return fmt.Errorf("gagal memeriksa referensi: %w", err)
	}
	if referenced {
		return fmt.Errorf("jurusan tidak dapat dihapus karena masih digunakan oleh kelas")
	}

	return s.repo.Delete(sekolahID, id)
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
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data jurusan")
		return
	}
	response.JSON(w, 200, list)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	var req CreateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}
	j, err := h.service.Create(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}
	response.JSON(w, 201, j)
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
	j, err := h.service.Update(user.SekolahID, id, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, 200, j)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}
	if err := h.service.Delete(user.SekolahID, id); err != nil {
		response.Error(w, 422, "DELETE_FAILED", err.Error())
		return
	}
	response.JSON(w, 200, map[string]string{"message": "Jurusan berhasil dihapus"})
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/jurusan", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}
