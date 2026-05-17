package notifikasi

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type Preferensi struct {
	ID             int64  `json:"id"`
	SekolahID      int64  `json:"sekolah_id"`
	PenggunaID     *int64 `json:"pengguna_id"`
	SiswaID        *int64 `json:"siswa_id"`
	RecipientType  string `json:"recipient_type"`
	Channel        string `json:"channel"`
	Destination    string `json:"destination"`
	Enabled        bool   `json:"enabled"`
	ConsentStatus  string `json:"consent_status"`
	ConsentSource  string `json:"consent_source"`
	ConsentAt      string `json:"consent_at"`
	RevokedAt      string `json:"revoked_at"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type PreferensiListParams struct {
	Page          int
	Limit         int
	Channel       string
	ConsentStatus string
	Enabled       string
}

type PreferensiRepository struct {
	db *sql.DB
}

func NewPreferensiRepository(db *sql.DB) *PreferensiRepository {
	return &PreferensiRepository{db: db}
}

func (r *PreferensiRepository) List(sekolahID int64, params PreferensiListParams) ([]Preferensi, int, error) {
	query := sq.Select("id", "sekolah_id",
		"COALESCE(pengguna_id,0)", "COALESCE(siswa_id,0)",
		"recipient_type", "channel", "destination",
		"enabled", "consent_status", "consent_source",
		"COALESCE(consent_at,'')", "COALESCE(revoked_at,'')",
		"created_at", "COALESCE(updated_at,created_at)").
		From("notifikasi_preferensi").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("notifikasi_preferensi").
		Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Channel != "" {
		query = query.Where(sq.Eq{"channel": params.Channel})
		countQuery = countQuery.Where(sq.Eq{"channel": params.Channel})
	}
	if params.ConsentStatus != "" {
		query = query.Where(sq.Eq{"consent_status": params.ConsentStatus})
		countQuery = countQuery.Where(sq.Eq{"consent_status": params.ConsentStatus})
	}
	if params.Enabled != "" {
		if params.Enabled == "true" {
			query = query.Where(sq.Eq{"enabled": true})
			countQuery = countQuery.Where(sq.Eq{"enabled": true})
		} else {
			query = query.Where(sq.Eq{"enabled": false})
			countQuery = countQuery.Where(sq.Eq{"enabled": false})
		}
	}

	var total int
	if err := countQuery.RunWith(r.db).QueryRow().Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query = query.OrderBy("updated_at DESC").Limit(uint64(params.Limit)).Offset(uint64(offset))

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Preferensi
	for rows.Next() {
		var p Preferensi
		var penggunaID, siswaID int64
		if err := rows.Scan(&p.ID, &p.SekolahID, &penggunaID, &siswaID,
			&p.RecipientType, &p.Channel, &p.Destination,
			&p.Enabled, &p.ConsentStatus, &p.ConsentSource,
			&p.ConsentAt, &p.RevokedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if penggunaID != 0 {
			p.PenggunaID = &penggunaID
		}
		if siswaID != 0 {
			p.SiswaID = &siswaID
		}
		list = append(list, p)
	}
	return list, total, nil
}

func (r *PreferensiRepository) Upsert(p *Preferensi) (int64, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	var consentAt, revokedAt interface{}
	if p.ConsentStatus == "granted" {
		consentAt = now
	}
	if p.ConsentStatus == "revoked" {
		revokedAt = now
	}

	result, err := sq.Insert("notifikasi_preferensi").
		Columns("sekolah_id", "pengguna_id", "siswa_id", "recipient_type",
			"channel", "destination", "enabled", "consent_status", "consent_source",
			"consent_at", "revoked_at", "updated_at").
		Values(p.SekolahID, p.PenggunaID, p.SiswaID, p.RecipientType,
			p.Channel, p.Destination, p.Enabled, p.ConsentStatus, p.ConsentSource,
			consentAt, revokedAt, now).
		Suffix(`ON CONFLICT(sekolah_id, channel, destination) DO UPDATE SET
			enabled = excluded.enabled,
			consent_status = excluded.consent_status,
			consent_source = excluded.consent_source,
			consent_at = CASE WHEN excluded.consent_status = 'granted' THEN excluded.consent_at ELSE notifikasi_preferensi.consent_at END,
			revoked_at = CASE WHEN excluded.consent_status = 'revoked' THEN excluded.revoked_at ELSE notifikasi_preferensi.revoked_at END,
			updated_at = excluded.updated_at`).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *PreferensiRepository) CanSend(sekolahID int64, channel, destination string) (bool, string) {
	var enabled bool
	var consentStatus string
	err := sq.Select("enabled", "consent_status").
		From("notifikasi_preferensi").
		Where(sq.Eq{
			"sekolah_id":  sekolahID,
			"channel":     channel,
			"destination": destination,
		}).
		RunWith(r.db).QueryRow().Scan(&enabled, &consentStatus)

	if err == sql.ErrNoRows {
		return false, "no preference found"
	}
	if err != nil {
		return false, fmt.Sprintf("lookup error: %v", err)
	}
	if !enabled {
		return false, "channel disabled"
	}
	if consentStatus != "granted" {
		return false, fmt.Sprintf("consent %s", consentStatus)
	}
	return true, ""
}

type PreferensiService struct {
	repo *PreferensiRepository
}

func NewPreferensiService(repo *PreferensiRepository) *PreferensiService {
	return &PreferensiService{repo: repo}
}

type UpsertPreferensiRequest struct {
	RecipientType string `json:"recipient_type"`
	Channel       string `json:"channel"`
	Destination   string `json:"destination"`
	Enabled       *bool  `json:"enabled"`
	ConsentStatus string `json:"consent_status"`
	ConsentSource string `json:"consent_source"`
}

func (s *PreferensiService) List(sekolahID int64, params PreferensiListParams) ([]Preferensi, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *PreferensiService) Upsert(sekolahID int64, req UpsertPreferensiRequest) (*Preferensi, error) {
	errs := validator.Collect(
		validator.Required("channel", req.Channel),
		validator.Required("destination", req.Destination),
		validator.Required("consent_status", req.ConsentStatus),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	if ve := validator.InList("consent_status", req.ConsentStatus, []string{"pending", "granted", "revoked"}); ve != nil {
		return nil, validator.ValidationErrors{*ve}
	}

	recipientType := req.RecipientType
	if recipientType == "" {
		recipientType = "manual"
	}
	consentSource := req.ConsentSource
	if consentSource == "" {
		consentSource = "admin"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	p := &Preferensi{
		SekolahID:     sekolahID,
		RecipientType: recipientType,
		Channel:       req.Channel,
		Destination:   req.Destination,
		Enabled:       enabled,
		ConsentStatus: req.ConsentStatus,
		ConsentSource: consentSource,
	}

	if _, err := s.repo.Upsert(p); err != nil {
		return nil, fmt.Errorf("gagal menyimpan preferensi: %w", err)
	}

	return p, nil
}

type PreferensiHandler struct {
	service *PreferensiService
}

func NewPreferensiHandler(service *PreferensiService) *PreferensiHandler {
	return &PreferensiHandler{service: service}
}

func (h *PreferensiHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parsePreferensiListParams(r)

	list, total, err := h.service.List(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil preferensi")
		return
	}

	totalPages := (total + params.Limit - 1) / params.Limit
	response.JSONWithMeta(w, 200, list, &response.Meta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *PreferensiHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req UpsertPreferensiRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	p, err := h.service.Upsert(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, p)
}

func parsePreferensiListParams(r *http.Request) PreferensiListParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return PreferensiListParams{
		Page:          page,
		Limit:         limit,
		Channel:       r.URL.Query().Get("channel"),
		ConsentStatus: r.URL.Query().Get("consent_status"),
		Enabled:       r.URL.Query().Get("enabled"),
	}
}

func RegisterPreferensiRoutes(r chi.Router, h *PreferensiHandler) {
	r.Route("/notifikasi/preferensi", func(r chi.Router) {
		r.Get("/", h.List)
		r.Put("/", h.Upsert)
	})
}
