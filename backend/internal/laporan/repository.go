package laporan

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

type DashboardStats struct {
	TotalSiswaAktif         int     `json:"total_siswa_aktif"`
	TotalPembayaranBulanIni float64 `json:"total_pembayaran_bulan_ini"`
	PendaftarPPDBBaru       int     `json:"pendaftar_ppdb_baru"`
	TagihanJatuhTempo       int     `json:"tagihan_jatuh_tempo"`
	PembayaranPending       int     `json:"pembayaran_pending"`
	TotalKelas              int     `json:"total_kelas"`
}

type RekapPembayaranItem struct {
	Tanggal        string  `json:"tanggal"`
	Metode         string  `json:"metode"`
	TotalTransaksi int     `json:"total_transaksi"`
	TotalNominal   float64 `json:"total_nominal"`
}

type RekapPPDB struct {
	TotalPendaftar int `json:"total_pendaftar"`
	Menunggu       int `json:"menunggu"`
	BerkasLengkap  int `json:"berkas_lengkap"`
	Diterima       int `json:"diterima"`
	TidakDiterima  int `json:"tidak_diterima"`
	Cadangan       int `json:"cadangan"`
	DaftarUlang    int `json:"daftar_ulang"`
}

type RekapSiswa struct {
	Total     int `json:"total"`
	Aktif     int `json:"aktif"`
	Lulus     int `json:"lulus"`
	Pindah    int `json:"pindah"`
	Keluar    int `json:"keluar"`
	LakiLaki  int `json:"laki_laki"`
	Perempuan int `json:"perempuan"`
}

func (r *Repository) GetDashboardStats(sekolahID int64) (*DashboardStats, error) {
	stats := &DashboardStats{}

	err := sq.Select("COUNT(*)").From("siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "status": "aktif"}).
		RunWith(r.db).QueryRow().Scan(&stats.TotalSiswaAktif)
	if err != nil {
		return nil, err
	}

	var totalPembayaran sql.NullFloat64
	err = r.db.QueryRow(
		"SELECT COALESCE(SUM(p.jumlah),0) FROM pembayaran p JOIN tagihan t ON p.tagihan_id = t.id WHERE t.sekolah_id = ? AND p.status = 'verified' AND strftime('%Y-%m', p.tanggal) = strftime('%Y-%m', 'now')",
		sekolahID).Scan(&totalPembayaran)
	if err != nil {
		return nil, err
	}
	stats.TotalPembayaranBulanIni = totalPembayaran.Float64

	err = sq.Select("COUNT(*)").From("ppdb_pendaftaran").
		Where(sq.Eq{"sekolah_id": sekolahID, "status": "menunggu"}).
		RunWith(r.db).QueryRow().Scan(&stats.PendaftarPPDBBaru)
	if err != nil {
		return nil, err
	}

	err = sq.Select("COUNT(*)").From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		Where(sq.NotEq{"status": "lunas"}).
		Where("jatuh_tempo <= date('now')").
		RunWith(r.db).QueryRow().Scan(&stats.TagihanJatuhTempo)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM pembayaran p JOIN tagihan t ON p.tagihan_id = t.id WHERE t.sekolah_id = ? AND p.status = 'pending'",
		sekolahID).Scan(&stats.PembayaranPending)
	if err != nil {
		return nil, err
	}

	err = sq.Select("COUNT(*)").From("kelas").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().Scan(&stats.TotalKelas)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *Repository) GetRekapPembayaran(sekolahID int64, tanggalMulai, tanggalSelesai string) ([]RekapPembayaranItem, error) {
	rows, err := r.db.Query(
		"SELECT date(p.tanggal) as tgl, p.metode, COUNT(*) as total_transaksi, SUM(p.jumlah) as total_nominal FROM pembayaran p JOIN tagihan t ON p.tagihan_id = t.id WHERE t.sekolah_id = ? AND p.status = 'verified' AND p.tanggal >= ? AND p.tanggal <= ? GROUP BY tgl, p.metode ORDER BY tgl DESC",
		sekolahID, tanggalMulai, tanggalSelesai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []RekapPembayaranItem
	for rows.Next() {
		var item RekapPembayaranItem
		if err := rows.Scan(&item.Tanggal, &item.Metode, &item.TotalTransaksi, &item.TotalNominal); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *Repository) GetRekapPPDB(sekolahID int64, tahunAjaranID int64) (*RekapPPDB, error) {
	rows, err := sq.Select("status", "COUNT(*)").
		From("ppdb_pendaftaran").
		Where(sq.Eq{"sekolah_id": sekolahID, "tahun_ajaran_id": tahunAjaranID}).
		GroupBy("status").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rekap := &RekapPPDB{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		rekap.TotalPendaftar += count
		switch status {
		case "menunggu":
			rekap.Menunggu = count
		case "berkas_lengkap":
			rekap.BerkasLengkap = count
		case "diterima":
			rekap.Diterima = count
		case "tidak_diterima":
			rekap.TidakDiterima = count
		case "cadangan":
			rekap.Cadangan = count
		case "daftar_ulang":
			rekap.DaftarUlang = count
		}
	}
	return rekap, nil
}

func (r *Repository) GetRekapSiswa(sekolahID int64) (*RekapSiswa, error) {
	rekap := &RekapSiswa{}

	rows, err := sq.Select("status", "COUNT(*)").
		From("siswa").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		GroupBy("status").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		rekap.Total += count
		switch status {
		case "aktif":
			rekap.Aktif = count
		case "lulus":
			rekap.Lulus = count
		case "pindah":
			rekap.Pindah = count
		case "keluar":
			rekap.Keluar = count
		}
	}

	genderRows, err := sq.Select("jenis_kelamin", "COUNT(*)").
		From("siswa").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		GroupBy("jenis_kelamin").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer genderRows.Close()

	for genderRows.Next() {
		var jk string
		var count int
		if err := genderRows.Scan(&jk, &count); err != nil {
			return nil, err
		}
		switch jk {
		case "L":
			rekap.LakiLaki = count
		case "P":
			rekap.Perempuan = count
		}
	}

	return rekap, nil
}
