package user

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type Service interface {
	GetUsersList(ctx context.Context, queryParam url.Values) pkg.Response
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{
		repo: r,
	}
}

func (s *service) GetUsersList(ctx context.Context, queryParam url.Values) pkg.Response {
	return pkg.NewResponse(http.StatusOK, "", nil, nil)
}
