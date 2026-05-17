package ppdb

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

type Pendaftaran struct {
	ID            int64   `json:"id"`
	SekolahID     int64   `json:"sekolah_id"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	NamaLengkap   string  `json:"nama_lengkap"`
	NIK           string  `json:"nik"`
	TempatLahir   string  `json:"tempat_lahir"`
	TanggalLahir  string  `json:"tanggal_lahir"`
	JenisKelamin  string  `json:"jenis_kelamin"`
	Agama         string  `json:"agama"`
	Alamat        string  `json:"alamat"`
	AsalSekolah   string  `json:"asal_sekolah"`
	NoHP          string  `json:"no_hp"`
	Email         string  `json:"email"`
	NamaOrtu      string  `json:"nama_ortu"`
	NoHPOrtu      string  `json:"no_hp_ortu"`
	PekerjaanOrtu string  `json:"pekerjaan_ortu"`
	Foto          string  `json:"foto"`
	Status        string  `json:"status"`
	Skor          float64 `json:"skor"`
	Ranking       int     `json:"ranking"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Catatan       string  `json:"catatan"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type Berkas struct {
	ID            int64  `json:"id"`
	PendaftaranID int64  `json:"pendaftaran_id"`
	JenisBerkas   string `json:"jenis_berkas"`
	FilePath      string `json:"file_path"`
	Status        string `json:"status"`
	Catatan       string `json:"catatan"`
	CreatedAt     string `json:"created_at"`
}

type Ujian struct {
	ID            int64   `json:"id"`
	PendaftaranID int64   `json:"pendaftaran_id"`
	NamaUjian     string  `json:"nama_ujian"`
	Nilai         float64 `json:"nilai"`
	Keterangan    string  `json:"keterangan"`
	CreatedAt     string  `json:"created_at"`
}

type Pengumuman struct {
	ID                 int64  `json:"id"`
	PendaftaranID      int64  `json:"pendaftaran_id"`
	Status             string `json:"status"`
	Ranking            int    `json:"ranking"`
	Keterangan         string `json:"keterangan"`
	TanggalPengumuman  string `json:"tanggal_pengumuman"`
	CreatedAt          string `json:"created_at"`
}

type KonfigurasiRanking struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Metode        string `json:"metode"`
	BobotJSON     string `json:"bobot_json"`
	Kuota         int    `json:"kuota"`
	Cadangan      int    `json:"cadangan"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ListParams struct {
	Page          int
	Limit         int
	Status        string
	TahunAjaranID int64
	DaftarUlang   string
}

type RankingLog struct {
	ID                 int64  `json:"id"`
	SekolahID          int64  `json:"sekolah_id"`
	TahunAjaranID      int64  `json:"tahun_ajaran_id"`
	Metode             string `json:"metode"`
	BobotJSON          string `json:"bobot_json"`
	Kuota              int    `json:"kuota"`
	Cadangan           int    `json:"cadangan"`
	TotalPendaftar     int    `json:"total_pendaftar"`
	DiterimaCount      int    `json:"diterima_count"`
	CadanganCount      int    `json:"cadangan_count"`
	TidakDiterimaCount int    `json:"tidak_diterima_count"`
	DryRun             bool   `json:"dry_run"`
	ExecutedBy         int64  `json:"executed_by"`
	ExecutedAt         string `json:"executed_at"`
}

type RankedPendaftaran struct {
	ID           int64   `json:"id"`
	NamaLengkap  string  `json:"nama_lengkap"`
	Skor         float64 `json:"skor"`
	Ranking      int     `json:"ranking"`
	Status       string  `json:"status"`
	TanggalLahir string  `json:"tanggal_lahir"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

func (r *Repository) ListPendaftaran(sekolahID int64, params ListParams) ([]Pendaftaran, int, error) {
	query := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama_lengkap",
		"COALESCE(nik,'')", "COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')",
		"jenis_kelamin", "COALESCE(agama,'')", "COALESCE(alamat,'')",
		"COALESCE(asal_sekolah,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(pekerjaan_ortu,'')",
		"COALESCE(foto,'')", "status", "COALESCE(skor,0)", "COALESCE(ranking,0)",
		"COALESCE(latitude,0)", "COALESCE(longitude,0)", "COALESCE(catatan,'')",
		"created_at", "updated_at").
		From("ppdb_pendaftaran").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("ppdb_pendaftaran").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Status != "" {
		query = query.Where(sq.Eq{"status": params.Status})
		countQuery = countQuery.Where(sq.Eq{"status": params.Status})
	}
	if params.TahunAjaranID > 0 {
		query = query.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
		countQuery = countQuery.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
	}
	if params.DaftarUlang != "" {
		query = query.Where(sq.Eq{"daftar_ulang_status": params.DaftarUlang})
		countQuery = countQuery.Where(sq.Eq{"daftar_ulang_status": params.DaftarUlang})
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

	var list []Pendaftaran
	for rows.Next() {
		var p Pendaftaran
		err := rows.Scan(&p.ID, &p.SekolahID, &p.TahunAjaranID, &p.NamaLengkap,
			&p.NIK, &p.TempatLahir, &p.TanggalLahir, &p.JenisKelamin, &p.Agama,
			&p.Alamat, &p.AsalSekolah, &p.NoHP, &p.Email, &p.NamaOrtu, &p.NoHPOrtu,
			&p.PekerjaanOrtu, &p.Foto, &p.Status, &p.Skor, &p.Ranking,
			&p.Latitude, &p.Longitude, &p.Catatan, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, nil
}

func (r *Repository) GetPendaftaranByID(sekolahID, id int64) (*Pendaftaran, error) {
	var p Pendaftaran
	err := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama_lengkap",
		"COALESCE(nik,'')", "COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')",
		"jenis_kelamin", "COALESCE(agama,'')", "COALESCE(alamat,'')",
		"COALESCE(asal_sekolah,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(pekerjaan_ortu,'')",
		"COALESCE(foto,'')", "status", "COALESCE(skor,0)", "COALESCE(ranking,0)",
		"COALESCE(latitude,0)", "COALESCE(longitude,0)", "COALESCE(catatan,'')",
		"created_at", "updated_at").
		From("ppdb_pendaftaran").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&p.ID, &p.SekolahID, &p.TahunAjaranID, &p.NamaLengkap,
			&p.NIK, &p.TempatLahir, &p.TanggalLahir, &p.JenisKelamin, &p.Agama,
			&p.Alamat, &p.AsalSekolah, &p.NoHP, &p.Email, &p.NamaOrtu, &p.NoHPOrtu,
			&p.PekerjaanOrtu, &p.Foto, &p.Status, &p.Skor, &p.Ranking,
			&p.Latitude, &p.Longitude, &p.Catatan, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreatePendaftaran(p *Pendaftaran) (int64, error) {
	result, err := sq.Insert("ppdb_pendaftaran").
		Columns("sekolah_id", "tahun_ajaran_id", "nama_lengkap", "nik", "tempat_lahir",
			"tanggal_lahir", "jenis_kelamin", "agama", "alamat", "asal_sekolah",
			"no_hp", "email", "nama_ortu", "no_hp_ortu", "pekerjaan_ortu",
			"foto", "status", "latitude", "longitude").
		Values(p.SekolahID, p.TahunAjaranID, p.NamaLengkap, p.NIK, p.TempatLahir,
			p.TanggalLahir, p.JenisKelamin, p.Agama, p.Alamat, p.AsalSekolah,
			p.NoHP, p.Email, p.NamaOrtu, p.NoHPOrtu, p.PekerjaanOrtu,
			p.Foto, p.Status, p.Latitude, p.Longitude).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdatePendaftaranStatus(sekolahID, id int64, status string) error {
	_, err := sq.Update("ppdb_pendaftaran").
		Set("status", status).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) GetSekolahIDByTahunAjaran(tahunAjaranID int64) (int64, error) {
	var sekolahID int64
	err := sq.Select("sekolah_id").From("tahun_ajaran").
		Where(sq.Eq{"id": tahunAjaranID}).
		RunWith(r.db).QueryRow().Scan(&sekolahID)
	if err != nil {
		return 0, err
	}
	return sekolahID, nil
}

func (r *Repository) ListBerkasByPendaftaran(pendaftaranID int64) ([]Berkas, error) {
	rows, err := sq.Select("id", "pendaftaran_id", "jenis_berkas", "file_path",
		"status", "COALESCE(catatan,'')", "created_at").
		From("ppdb_berkas").
		Where(sq.Eq{"pendaftaran_id": pendaftaranID}).
		OrderBy("created_at ASC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Berkas
	for rows.Next() {
		var b Berkas
		err := rows.Scan(&b.ID, &b.PendaftaranID, &b.JenisBerkas, &b.FilePath,
			&b.Status, &b.Catatan, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func (r *Repository) CreateBerkas(b *Berkas) (int64, error) {
	result, err := sq.Insert("ppdb_berkas").
		Columns("pendaftaran_id", "jenis_berkas", "file_path", "status", "catatan").
		Values(b.PendaftaranID, b.JenisBerkas, b.FilePath, b.Status, b.Catatan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateBerkasStatus(sekolahID, id int64, status, catatan string) error {
	_, err := r.db.Exec(
		"UPDATE ppdb_berkas SET status=?, catatan=? WHERE id=? AND pendaftaran_id IN (SELECT id FROM ppdb_pendaftaran WHERE sekolah_id=?)",
		status, catatan, id, sekolahID)
	return err
}

func (r *Repository) ListUjianByPendaftaran(pendaftaranID int64) ([]Ujian, error) {
	rows, err := sq.Select("id", "pendaftaran_id", "nama_ujian",
		"COALESCE(nilai,0)", "COALESCE(keterangan,'')", "created_at").
		From("ppdb_ujian").
		Where(sq.Eq{"pendaftaran_id": pendaftaranID}).
		OrderBy("created_at ASC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Ujian
	for rows.Next() {
		var u Ujian
		err := rows.Scan(&u.ID, &u.PendaftaranID, &u.NamaUjian, &u.Nilai, &u.Keterangan, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

func (r *Repository) CreateUjian(u *Ujian) (int64, error) {
	result, err := sq.Insert("ppdb_ujian").
		Columns("pendaftaran_id", "nama_ujian", "nilai", "keterangan").
		Values(u.PendaftaranID, u.NamaUjian, u.Nilai, u.Keterangan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetPengumumanByPendaftaranID(pendaftaranID int64) (*Pengumuman, error) {
	var p Pengumuman
	err := sq.Select("id", "pendaftaran_id", "status", "COALESCE(ranking,0)",
		"COALESCE(keterangan,'')", "COALESCE(tanggal_pengumuman,'')", "created_at").
		From("ppdb_pengumuman").
		Where(sq.Eq{"pendaftaran_id": pendaftaranID}).
		OrderBy("created_at DESC").
		Limit(1).
		RunWith(r.db).QueryRow().
		Scan(&p.ID, &p.PendaftaranID, &p.Status, &p.Ranking, &p.Keterangan, &p.TanggalPengumuman, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreatePengumuman(p *Pengumuman) (int64, error) {
	result, err := sq.Insert("ppdb_pengumuman").
		Columns("pendaftaran_id", "status", "ranking", "keterangan", "tanggal_pengumuman").
		Values(p.PendaftaranID, p.Status, p.Ranking, p.Keterangan, p.TanggalPengumuman).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetKonfigurasiRanking(sekolahID, tahunAjaranID int64) (*KonfigurasiRanking, error) {
	var k KonfigurasiRanking
	err := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "metode",
		"COALESCE(bobot_json,'')", "kuota", "COALESCE(cadangan,0)", "created_at", "updated_at").
		From("ppdb_konfigurasi_ranking").
		Where(sq.Eq{"sekolah_id": sekolahID, "tahun_ajaran_id": tahunAjaranID}).
		RunWith(r.db).QueryRow().
		Scan(&k.ID, &k.SekolahID, &k.TahunAjaranID, &k.Metode, &k.BobotJSON,
			&k.Kuota, &k.Cadangan, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *Repository) UpsertKonfigurasiRanking(k *KonfigurasiRanking) error {
	existing, _ := r.GetKonfigurasiRanking(k.SekolahID, k.TahunAjaranID)
	if existing != nil {
		_, err := sq.Update("ppdb_konfigurasi_ranking").
			Set("metode", k.Metode).
			Set("bobot_json", k.BobotJSON).
			Set("kuota", k.Kuota).
			Set("cadangan", k.Cadangan).
			Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
			Where(sq.Eq{"id": existing.ID}).
			RunWith(r.db).Exec()
		return err
	}
	_, err := sq.Insert("ppdb_konfigurasi_ranking").
		Columns("sekolah_id", "tahun_ajaran_id", "metode", "bobot_json", "kuota", "cadangan").
		Values(k.SekolahID, k.TahunAjaranID, k.Metode, k.BobotJSON, k.Kuota, k.Cadangan).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ListAllPendaftaran(sekolahID int64, params ListParams) ([]Pendaftaran, error) {
	query := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama_lengkap",
		"COALESCE(nik,'')", "COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')",
		"jenis_kelamin", "COALESCE(agama,'')", "COALESCE(alamat,'')",
		"COALESCE(asal_sekolah,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(pekerjaan_ortu,'')",
		"COALESCE(foto,'')", "status", "COALESCE(skor,0)", "COALESCE(ranking,0)",
		"COALESCE(latitude,0)", "COALESCE(longitude,0)", "COALESCE(catatan,'')",
		"created_at", "updated_at").
		From("ppdb_pendaftaran").
		Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Status != "" {
		query = query.Where(sq.Eq{"status": params.Status})
	}
	if params.TahunAjaranID > 0 {
		query = query.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
	}

	query = query.OrderBy("created_at DESC")

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Pendaftaran
	for rows.Next() {
		var p Pendaftaran
		err := rows.Scan(&p.ID, &p.SekolahID, &p.TahunAjaranID, &p.NamaLengkap,
			&p.NIK, &p.TempatLahir, &p.TanggalLahir, &p.JenisKelamin, &p.Agama,
			&p.Alamat, &p.AsalSekolah, &p.NoHP, &p.Email, &p.NamaOrtu, &p.NoHPOrtu,
			&p.PekerjaanOrtu, &p.Foto, &p.Status, &p.Skor, &p.Ranking,
			&p.Latitude, &p.Longitude, &p.Catatan, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) GetUjianAvgByPendaftaran(pendaftaranID int64) (float64, error) {
	var avg sql.NullFloat64
	err := sq.Select("COALESCE(AVG(nilai), 0)").
		From("ppdb_ujian").
		Where(sq.Eq{"pendaftaran_id": pendaftaranID}).
		RunWith(r.db).QueryRow().Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg.Float64, nil
}

func (r *Repository) GetAllPendaftaranForRanking(sekolahID, tahunAjaranID int64) ([]Pendaftaran, error) {
	rows, err := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama_lengkap",
		"COALESCE(nik,'')", "COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')",
		"jenis_kelamin", "COALESCE(agama,'')", "COALESCE(alamat,'')",
		"COALESCE(asal_sekolah,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(pekerjaan_ortu,'')",
		"COALESCE(foto,'')", "status", "COALESCE(skor,0)", "COALESCE(ranking,0)",
		"COALESCE(latitude,0)", "COALESCE(longitude,0)", "COALESCE(catatan,'')",
		"created_at", "updated_at").
		From("ppdb_pendaftaran").
		Where(sq.Eq{"sekolah_id": sekolahID, "tahun_ajaran_id": tahunAjaranID}).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Pendaftaran
	for rows.Next() {
		var p Pendaftaran
		err := rows.Scan(&p.ID, &p.SekolahID, &p.TahunAjaranID, &p.NamaLengkap,
			&p.NIK, &p.TempatLahir, &p.TanggalLahir, &p.JenisKelamin, &p.Agama,
			&p.Alamat, &p.AsalSekolah, &p.NoHP, &p.Email, &p.NamaOrtu, &p.NoHPOrtu,
			&p.PekerjaanOrtu, &p.Foto, &p.Status, &p.Skor, &p.Ranking,
			&p.Latitude, &p.Longitude, &p.Catatan, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) UpdatePendaftaranSkorRankingStatus(sekolahID, id int64, skor float64, ranking int, status string) error {
	_, err := sq.Update("ppdb_pendaftaran").
		Set("skor", skor).
		Set("ranking", ranking).
		Set("status", status).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ResetRankingForTahunAjaran(sekolahID, tahunAjaranID int64) error {
	_, err := sq.Update("ppdb_pendaftaran").
		Set("skor", 0).
		Set("ranking", 0).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"sekolah_id": sekolahID, "tahun_ajaran_id": tahunAjaranID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) CreateRankingLog(log *RankingLog) (int64, error) {
	dryRun := 0
	if log.DryRun {
		dryRun = 1
	}
	result, err := sq.Insert("ppdb_ranking_log").
		Columns("sekolah_id", "tahun_ajaran_id", "metode", "bobot_json", "kuota", "cadangan",
			"total_pendaftar", "diterima_count", "cadangan_count", "tidak_diterima_count",
			"dry_run", "executed_by").
		Values(log.SekolahID, log.TahunAjaranID, log.Metode, log.BobotJSON, log.Kuota, log.Cadangan,
			log.TotalPendaftar, log.DiterimaCount, log.CadanganCount, log.TidakDiterimaCount,
			dryRun, log.ExecutedBy).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateDaftarUlangStatus(sekolahID, id int64, status string) error {
	var waktu interface{}
	if status == "sudah" {
		waktu = sq.Expr("datetime('now')")
	}
	_, err := sq.Update("ppdb_pendaftaran").
		Set("daftar_ulang_status", status).
		Set("daftar_ulang_at", waktu).
		Set("status", "daftar_ulang").
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) CreateBulkPengumuman(pengumuman []Pengumuman) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, p := range pengumuman {
		_, err := sq.Insert("ppdb_pengumuman").
			Columns("pendaftaran_id", "status", "ranking", "keterangan", "tanggal_pengumuman").
			Values(p.PendaftaranID, p.Status, p.Ranking, p.Keterangan, p.TanggalPengumuman).
			RunWith(tx).Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
