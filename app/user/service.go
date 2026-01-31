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
	GetUserDetail(ctx context.Context, userID string) pkg.Response
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

func (s *service) GetUserDetail(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FetchOneUser(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	// Map to UserProfile response
	userProfile := UserProfile{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return pkg.NewResponse(http.StatusOK, "User details retrieved successfully", nil, userProfile)
}
