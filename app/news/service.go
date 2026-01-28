package news

import "time"

type Service interface {
	FetchAllNews() ([]News, error)
	Create(news *News) error
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

func (s *service) FetchAllNews() ([]News, error) {
	return s.repo.FetchAllNews()
}

func (s *service) Create(news *News) error {
	return s.repo.Create(news)
}
