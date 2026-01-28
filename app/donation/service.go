package donation

import "time"

type Service interface {
	FetchAllDonations() ([]Donation, error)
	Create(donation *Donation) error
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *service) FetchAllDonations() ([]Donation, error) {
	return s.repo.FetchAllDonations()
}
func (s *service) Create(donation *Donation) error {
	return s.repo.Create(donation)
}
