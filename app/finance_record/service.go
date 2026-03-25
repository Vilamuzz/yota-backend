package finance_record

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type Service interface {
	CreateRecord(ctx context.Context, record *FinanceRecord) error
	ListRecords(ctx context.Context, params RecordQueryParams) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) CreateRecord(ctx context.Context, record *FinanceRecord) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	now := time.Now()
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.TransactionDate.IsZero() {
		record.TransactionDate = now
	}
	record.CreatedAt = now
	record.UpdatedAt = now

	return s.repo.Create(ctx, record)
}

func (s *service) ListRecords(ctx context.Context, params RecordQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	usingPrevCursor := params.PrevCursor != ""
	options := map[string]interface{}{
		"limit": params.Limit,
	}

	if params.FundID != "" {
		options["fund_id"] = params.FundID
	}
	if params.SourceType != "" {
		options["source_type"] = params.SourceType
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	records, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch finance records", nil, nil)
	}

	hasMore := len(records) > params.Limit
	if hasMore {
		records = records[:params.Limit]
	}

	// Reverse ASC → DESC when navigating backwards
	if usingPrevCursor {
		for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
			records[i], records[j] = records[j], records[i]
		}
	}

	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	var nextCursor, prevCursor string
	if hasNext && len(records) > 0 {
		last := records[len(records)-1]
		nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
	}
	if hasPrev && len(records) > 0 {
		first := records[0]
		prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
	}

	responses := make([]FinanceRecordResponse, len(records))
	for i, r := range records {
		rCopy := r
		responses[i] = toResponse(&rCopy)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, map[string]interface{}{
		"records": responses,
		"pagination": pkg.CursorPagination{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			Limit:      params.Limit,
		},
	})
}
