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
	ListReportedPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response
	DeletePrayer(ctx context.Context, id string) pkg.Response
	IncrementPrayerCount(ctx context.Context, id string) pkg.Response
	DecrementPrayerCount(ctx context.Context, id string) pkg.Response
	ReportPrayer(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) IncrementPrayerCount(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}
	prayer.LikeCount++
	if err := s.repo.Update(ctx, prayer); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer count incremented successfully", nil, nil)
}

func (s *service) DecrementPrayerCount(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}
	prayer.LikeCount--
	if err := s.repo.Update(ctx, prayer); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer count decremented successfully", nil, nil)
}

func (s *service) ReportPrayer(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}
	prayer.ReportCount++
	if err := s.repo.Update(ctx, prayer); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer reported successfully", nil, nil)
}

func (s *service) FindPrayerByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	prayer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer found successfully", nil, prayer.toPrayerResponse())
}

func (s *service) ListPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	prayers, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayers", nil, nil)
	}

	hasMore := len(prayers) > params.Limit
	if hasMore {
		prayers = prayers[:params.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(prayers)-1; i < j; i, j = i+1, j-1 {
			prayers[i], prayers[j] = prayers[j], prayers[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(prayers) > 0 {
		first := prayers[0]
		last := prayers[len(prayers)-1]
		if !usingPrevCursor {
			nextCursor = last.ID
			if hasMore {
				prevCursor = first.ID
			}
		} else {
			prevCursor = first.ID
			if hasMore {
				nextCursor = last.ID
			}
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Prayers found successfully", nil, toPrayerListResponse(prayers, pagination))
}

func (s *service) ListReportedPrayers(ctx context.Context, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit":    params.Limit,
		"reported": true,
	}
	if params.DonationID != "" {
		options["donation_id"] = params.DonationID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	prayers, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to find prayers", nil, nil)
	}

	hasMore := len(prayers) > params.Limit
	if hasMore {
		prayers = prayers[:params.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(prayers)-1; i < j; i, j = i+1, j-1 {
			prayers[i], prayers[j] = prayers[j], prayers[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(prayers) > 0 {
		first := prayers[0]
		last := prayers[len(prayers)-1]
		if !usingPrevCursor {
			nextCursor = last.ID
			if hasMore {
				prevCursor = first.ID
			}
		} else {
			prevCursor = first.ID
			if hasMore {
				nextCursor = last.ID
			}
		}
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Prayers found successfully", nil, toPrayerReportedListResponse(prayers, pagination))
}

func (s *service) DeletePrayer(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete prayer", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Prayer deleted successfully", nil, nil)
}
