package prayer

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type Service interface {
	FindPrayerByID(ctx context.Context, id string) pkg.Response
	ListPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) FindPrayerByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer found successfully", nil, prayer)
}

func (s *service) ListPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	options := make(map[string]interface{})
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.IsReported {
		options["is_reported"] = params.IsReported
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}
	if params.Limit > 0 {
		options["limit"] = params.Limit
	}

	prayers, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayers", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayers found successfully", nil, prayers)
}
