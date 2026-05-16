package notifikasi

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(sekolahID int64, params ListParams) ([]Notifikasi, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *Service) TestSend(sekolahID int64, tipe, penerima, pesan string) (*Notifikasi, error) {
	n := &Notifikasi{
		SekolahID:  sekolahID,
		Tipe:       tipe,
		Penerima:   penerima,
		Pesan:      pesan,
		Status:     "pending",
		MaxRetries: 3,
	}

	id, err := s.repo.Create(n)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateStatus(id, "sent", "", 0)

	n.ID = id
	n.Status = "sent"
	return n, nil
}

func (s *Service) GetQueueStats(sekolahID int64) (*QueueStats, error) {
	return s.repo.GetQueueStats(sekolahID)
}

func (s *Service) Enqueue(sekolahID int64, tipe, penerima, pesan string) (int64, error) {
	n := &Notifikasi{
		SekolahID:  sekolahID,
		Tipe:       tipe,
		Penerima:   penerima,
		Pesan:      pesan,
		Status:     "pending",
		MaxRetries: 3,
	}
	return s.repo.Create(n)
}
