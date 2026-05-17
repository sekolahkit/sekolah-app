package selfservice

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

type LinkedSiswa struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	NIS           string `json:"nis"`
	Nama          string `json:"nama"`
	JenisKelamin  string `json:"jenis_kelamin"`
	Status        string `json:"status"`
	Hubungan      string `json:"hubungan"`
}

type SiswaDetail struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	NIS           string `json:"nis"`
	Nama          string `json:"nama"`
	JenisKelamin  string `json:"jenis_kelamin"`
	TempatLahir   string `json:"tempat_lahir"`
	TanggalLahir  string `json:"tanggal_lahir"`
	Agama         string `json:"agama"`
	Alamat        string `json:"alamat"`
	NoHP          string `json:"no_hp"`
	Email         string `json:"email"`
	Status        string `json:"status"`
}

type Tagihan struct {
	ID            int64   `json:"id"`
	SiswaID       int64   `json:"siswa_id"`
	KategoriID    int64   `json:"kategori_id"`
	KategoriNama  string  `json:"kategori_nama"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	Semester      string  `json:"semester"`
	Nominal       float64 `json:"nominal"`
	JatuhTempo    string  `json:"jatuh_tempo"`
	Status        string  `json:"status"`
	Catatan       string  `json:"catatan"`
}

type Pembayaran struct {
	ID                int64   `json:"id"`
	TagihanID         int64   `json:"tagihan_id"`
	SiswaID           int64   `json:"siswa_id"`
	Jumlah            float64 `json:"jumlah"`
	Tanggal           string  `json:"tanggal"`
	Metode            string  `json:"metode"`
	BuktiBayar        string  `json:"bukti_bayar"`
	RekeningSekolahID int64   `json:"rekening_sekolah_id"`
	Status            string  `json:"status"`
	Catatan           string  `json:"catatan"`
	CreatedAt         string  `json:"created_at"`
}

type DashboardSiswa struct {
	TotalTagihan        int     `json:"total_tagihan"`
	TagihanBelumBayar   int     `json:"tagihan_belum_bayar"`
	TotalTerbayar       float64 `json:"total_terbayar"`
	PembayaranPending   int     `json:"pembayaran_pending"`
}

type DashboardOrangtua struct {
	JumlahAnak          int     `json:"jumlah_anak"`
	TotalTagihan        int     `json:"total_tagihan"`
	TagihanBelumBayar   int     `json:"tagihan_belum_bayar"`
	TotalTerbayar       float64 `json:"total_terbayar"`
	PembayaranPending   int     `json:"pembayaran_pending"`
}

type GuruKelas struct {
	ID              int64  `json:"id"`
	Nama            string `json:"nama"`
	Tingkat         int    `json:"tingkat"`
	JurusanNama     string `json:"jurusan_nama"`
	TahunAjaranNama string `json:"tahun_ajaran_nama"`
	JumlahSiswa     int    `json:"jumlah_siswa"`
}

type GuruSiswa struct {
	ID           int64  `json:"id"`
	NIS          string `json:"nis"`
	Nama         string `json:"nama"`
	JenisKelamin string `json:"jenis_kelamin"`
	Status       string `json:"status"`
}

type DashboardGuru struct {
	TotalKelas  int `json:"total_kelas"`
	TotalSiswa  int `json:"total_siswa"`
}

func (r *Repository) HasAccess(sekolahID, penggunaID, siswaID int64) (bool, error) {
	var count int
	err := sq.Select("COUNT(*)").From("pengguna_siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "pengguna_id": penggunaID, "siswa_id": siswaID, "aktif": true}).
		RunWith(r.db).QueryRow().Scan(&count)
	return count > 0, err
}

func (r *Repository) GetLinkedSiswa(sekolahID, penggunaID int64) ([]LinkedSiswa, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.sekolah_id, s.nis, s.nama, s.jenis_kelamin, s.status, ps.hubungan
		FROM pengguna_siswa ps
		JOIN siswa s ON ps.siswa_id = s.id
		WHERE ps.sekolah_id = ? AND ps.pengguna_id = ? AND ps.aktif = 1
		ORDER BY s.nama ASC`, sekolahID, penggunaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []LinkedSiswa
	for rows.Next() {
		var ls LinkedSiswa
		if err := rows.Scan(&ls.ID, &ls.SekolahID, &ls.NIS, &ls.Nama, &ls.JenisKelamin, &ls.Status, &ls.Hubungan); err != nil {
			return nil, err
		}
		result = append(result, ls)
	}
	return result, nil
}

func (r *Repository) GetSiswaDetail(sekolahID, siswaID int64) (*SiswaDetail, error) {
	s := &SiswaDetail{}
	var tempatLahir, tanggalLahir, agama, alamat, noHP, email sql.NullString
	err := r.db.QueryRow(`
		SELECT id, sekolah_id, nis, nama, jenis_kelamin, tempat_lahir, tanggal_lahir, agama, alamat, no_hp, email, status
		FROM siswa WHERE id = ? AND sekolah_id = ?`, siswaID, sekolahID).
		Scan(&s.ID, &s.SekolahID, &s.NIS, &s.Nama, &s.JenisKelamin, &tempatLahir, &tanggalLahir, &agama, &alamat, &noHP, &email, &s.Status)
	if err != nil {
		return nil, err
	}
	s.TempatLahir = tempatLahir.String
	s.TanggalLahir = tanggalLahir.String
	s.Agama = agama.String
	s.Alamat = alamat.String
	s.NoHP = noHP.String
	s.Email = email.String
	return s, nil
}

func (r *Repository) GetTagihanBySiswa(sekolahID, siswaID int64) ([]Tagihan, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.siswa_id, t.kategori_id, COALESCE(kp.nama,''), t.tahun_ajaran_id, COALESCE(t.semester,''), t.nominal, COALESCE(t.jatuh_tempo,''), t.status, COALESCE(t.catatan,'')
		FROM tagihan t
		LEFT JOIN kategori_pembayaran kp ON t.kategori_id = kp.id
		WHERE t.sekolah_id = ? AND t.siswa_id = ?
		ORDER BY t.jatuh_tempo DESC`, sekolahID, siswaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Tagihan
	for rows.Next() {
		var t Tagihan
		if err := rows.Scan(&t.ID, &t.SiswaID, &t.KategoriID, &t.KategoriNama, &t.TahunAjaranID, &t.Semester, &t.Nominal, &t.JatuhTempo, &t.Status, &t.Catatan); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *Repository) GetPembayaranBySiswa(sekolahID, siswaID int64) ([]Pembayaran, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.tagihan_id, p.siswa_id, p.jumlah, p.tanggal, p.metode, COALESCE(p.bukti_bayar,''), COALESCE(p.rekening_sekolah_id,0), p.status, COALESCE(p.catatan,''), p.created_at
		FROM pembayaran p
		JOIN tagihan t ON p.tagihan_id = t.id
		WHERE t.sekolah_id = ? AND p.siswa_id = ?
		ORDER BY p.created_at DESC`, sekolahID, siswaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Pembayaran
	for rows.Next() {
		var p Pembayaran
		if err := rows.Scan(&p.ID, &p.TagihanID, &p.SiswaID, &p.Jumlah, &p.Tanggal, &p.Metode, &p.BuktiBayar, &p.RekeningSekolahID, &p.Status, &p.Catatan, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *Repository) GetDashboardSiswa(sekolahID, siswaID int64) (*DashboardSiswa, error) {
	stats := &DashboardSiswa{}

	err := sq.Select("COUNT(*)").From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID, "siswa_id": siswaID}).
		RunWith(r.db).QueryRow().Scan(&stats.TotalTagihan)
	if err != nil {
		return nil, err
	}

	err = sq.Select("COUNT(*)").From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID, "siswa_id": siswaID}).
		Where(sq.NotEq{"status": "lunas"}).
		RunWith(r.db).QueryRow().Scan(&stats.TagihanBelumBayar)
	if err != nil {
		return nil, err
	}

	var total sql.NullFloat64
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(p.jumlah),0) FROM pembayaran p
		JOIN tagihan t ON p.tagihan_id = t.id
		WHERE t.sekolah_id = ? AND p.siswa_id = ? AND p.status = 'verified'`, sekolahID, siswaID).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats.TotalTerbayar = total.Float64

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM pembayaran p
		JOIN tagihan t ON p.tagihan_id = t.id
		WHERE t.sekolah_id = ? AND p.siswa_id = ? AND p.status = 'pending'`, sekolahID, siswaID).Scan(&stats.PembayaranPending)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *Repository) GetDashboardOrangtua(sekolahID, penggunaID int64) (*DashboardOrangtua, error) {
	stats := &DashboardOrangtua{}

	siswaIDs, err := r.getLinkedSiswaIDs(sekolahID, penggunaID)
	if err != nil {
		return nil, err
	}
	stats.JumlahAnak = len(siswaIDs)

	if len(siswaIDs) == 0 {
		return stats, nil
	}

	err = sq.Select("COUNT(*)").From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID, "siswa_id": siswaIDs}).
		RunWith(r.db).QueryRow().Scan(&stats.TotalTagihan)
	if err != nil {
		return nil, err
	}

	err = sq.Select("COUNT(*)").From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID, "siswa_id": siswaIDs}).
		Where(sq.NotEq{"status": "lunas"}).
		RunWith(r.db).QueryRow().Scan(&stats.TagihanBelumBayar)
	if err != nil {
		return nil, err
	}

	var total sql.NullFloat64
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(p.jumlah),0) FROM pembayaran p
		JOIN tagihan t ON p.tagihan_id = t.id
		WHERE t.sekolah_id = ? AND p.siswa_id IN (SELECT siswa_id FROM pengguna_siswa WHERE sekolah_id = ? AND pengguna_id = ? AND aktif = 1) AND p.status = 'verified'`,
		sekolahID, sekolahID, penggunaID).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats.TotalTerbayar = total.Float64

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM pembayaran p
		JOIN tagihan t ON p.tagihan_id = t.id
		WHERE t.sekolah_id = ? AND p.siswa_id IN (SELECT siswa_id FROM pengguna_siswa WHERE sekolah_id = ? AND pengguna_id = ? AND aktif = 1) AND p.status = 'pending'`,
		sekolahID, sekolahID, penggunaID).Scan(&stats.PembayaranPending)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *Repository) GetDashboardGuru(sekolahID, penggunaID int64) (*DashboardGuru, error) {
	stats := &DashboardGuru{}

	err := sq.Select("COUNT(*)").From("kelas").
		Where(sq.Eq{"sekolah_id": sekolahID, "wali_kelas_id": penggunaID}).
		RunWith(r.db).QueryRow().Scan(&stats.TotalKelas)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT ks.siswa_id) FROM kelas_siswa ks
		JOIN kelas k ON ks.kelas_id = k.id
		WHERE k.sekolah_id = ? AND k.wali_kelas_id = ?`, sekolahID, penggunaID).Scan(&stats.TotalSiswa)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *Repository) GetGuruKelas(sekolahID, penggunaID int64) ([]GuruKelas, error) {
	rows, err := r.db.Query(`
		SELECT k.id, k.nama, k.tingkat, COALESCE(j.nama,''), COALESCE(ta.nama,''),
			(SELECT COUNT(*) FROM kelas_siswa WHERE kelas_id = k.id)
		FROM kelas k
		LEFT JOIN jurusan j ON k.jurusan_id = j.id
		LEFT JOIN tahun_ajaran ta ON k.tahun_ajaran_id = ta.id
		WHERE k.sekolah_id = ? AND k.wali_kelas_id = ?
		ORDER BY ta.nama DESC, k.tingkat ASC, k.nama ASC`, sekolahID, penggunaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GuruKelas
	for rows.Next() {
		var gk GuruKelas
		if err := rows.Scan(&gk.ID, &gk.Nama, &gk.Tingkat, &gk.JurusanNama, &gk.TahunAjaranNama, &gk.JumlahSiswa); err != nil {
			return nil, err
		}
		result = append(result, gk)
	}
	return result, nil
}

func (r *Repository) GetGuruSiswaByKelas(sekolahID, penggunaID, kelasID int64) ([]GuruSiswa, error) {
	var count int
	err := sq.Select("COUNT(*)").From("kelas").
		Where(sq.Eq{"id": kelasID, "sekolah_id": sekolahID, "wali_kelas_id": penggunaID}).
		RunWith(r.db).QueryRow().Scan(&count)
	if err != nil || count == 0 {
		return nil, sql.ErrNoRows
	}

	rows, err := r.db.Query(`
		SELECT s.id, s.nis, s.nama, s.jenis_kelamin, s.status
		FROM kelas_siswa ks
		JOIN siswa s ON ks.siswa_id = s.id
		WHERE ks.kelas_id = ?
		ORDER BY s.nama ASC`, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GuruSiswa
	for rows.Next() {
		var gs GuruSiswa
		if err := rows.Scan(&gs.ID, &gs.NIS, &gs.Nama, &gs.JenisKelamin, &gs.Status); err != nil {
			return nil, err
		}
		result = append(result, gs)
	}
	return result, nil
}

func (r *Repository) TagihanBelongsToLinkedSiswa(sekolahID, penggunaID, tagihanID int64) (bool, int64, error) {
	var siswaID int64
	err := r.db.QueryRow(`
		SELECT t.siswa_id FROM tagihan t
		JOIN pengguna_siswa ps ON ps.siswa_id = t.siswa_id AND ps.sekolah_id = t.sekolah_id
		WHERE t.id = ? AND t.sekolah_id = ? AND ps.pengguna_id = ? AND ps.aktif = 1`,
		tagihanID, sekolahID, penggunaID).Scan(&siswaID)
	if err != nil {
		return false, 0, err
	}
	return true, siswaID, nil
}

func (r *Repository) CreatePembayaran(tagihanID, siswaID int64, jumlah float64, tanggal, metode, buktiBayar string, rekeningSekolahID int64, catatan string) (int64, error) {
	var providerVal, buktiVal, rekeningVal interface{}
	if buktiBayar != "" {
		buktiVal = buktiBayar
	}
	if rekeningSekolahID > 0 {
		rekeningVal = rekeningSekolahID
	}

	result, err := sq.Insert("pembayaran").
		Columns("tagihan_id", "siswa_id", "jumlah", "tanggal", "metode", "provider", "bukti_bayar", "rekening_sekolah_id", "catatan").
		Values(tagihanID, siswaID, jumlah, tanggal, metode, providerVal, buktiVal, rekeningVal, catatan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) getLinkedSiswaIDs(sekolahID, penggunaID int64) ([]int64, error) {
	rows, err := sq.Select("siswa_id").From("pengguna_siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "pengguna_id": penggunaID, "aktif": true}).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

var _ = time.Now
