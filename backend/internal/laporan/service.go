package laporan

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDashboard(sekolahID int64) (*DashboardStats, error) {
	return s.repo.GetDashboardStats(sekolahID)
}

func (s *Service) GetRekapPembayaran(sekolahID int64, tanggalMulai, tanggalSelesai string, tahunAjaranID int64) ([]RekapPembayaranItem, error) {
	return s.repo.GetRekapPembayaran(sekolahID, tanggalMulai, tanggalSelesai, tahunAjaranID)
}

func (s *Service) GetRekapPPDB(sekolahID int64, tahunAjaranID int64) (*RekapPPDB, error) {
	return s.repo.GetRekapPPDB(sekolahID, tahunAjaranID)
}

func (s *Service) GetRekapSiswa(sekolahID int64, tahunAjaranID int64) (*RekapSiswa, error) {
	return s.repo.GetRekapSiswa(sekolahID, tahunAjaranID)
}

func (s *Service) ExportPembayaran(sekolahID int64, tanggalMulai, tanggalSelesai string, tahunAjaranID int64) ([]ExportPembayaranRow, error) {
	return s.repo.ExportPembayaran(sekolahID, tanggalMulai, tanggalSelesai, tahunAjaranID)
}

func (s *Service) ExportPPDB(sekolahID int64, tahunAjaranID int64) ([]ExportPPDBRow, error) {
	return s.repo.ExportPPDB(sekolahID, tahunAjaranID)
}

func (s *Service) ExportSiswa(sekolahID int64, tahunAjaranID int64) ([]ExportSiswaRow, error) {
	return s.repo.ExportSiswa(sekolahID, tahunAjaranID)
}
