package user

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type Service interface {
	GetUsersList(ctx context.Context, queryParam url.Values) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(r Repository, timeout time.Duration) Service {
	return &service{
		repo:    r,
		timeout: timeout,
	}
}

func (s *service) GetUsersList(ctx context.Context, queryParam url.Values) pkg.Response {
	return pkg.NewResponse(http.StatusOK, "", nil, nil)
}
