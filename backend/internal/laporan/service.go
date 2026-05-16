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

func (s *Service) GetRekapPembayaran(sekolahID int64, tanggalMulai, tanggalSelesai string) ([]RekapPembayaranItem, error) {
	return s.repo.GetRekapPembayaran(sekolahID, tanggalMulai, tanggalSelesai)
}

func (s *Service) GetRekapPPDB(sekolahID int64, tahunAjaranID int64) (*RekapPPDB, error) {
	return s.repo.GetRekapPPDB(sekolahID, tahunAjaranID)
}

func (s *Service) GetRekapSiswa(sekolahID int64) (*RekapSiswa, error) {
	return s.repo.GetRekapSiswa(sekolahID)
}
