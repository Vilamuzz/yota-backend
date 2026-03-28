package foster_child

import "time"

type Service interface {
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
