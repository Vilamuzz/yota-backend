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
	GetSummary(ctx context.Context, isAdmin bool) pkg.Response
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

	return s.repo.Create(ctx, record)
}

func (s *service) GetSummary(ctx context.Context, isAdmin bool) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	summary, err := s.repo.Summary(ctx, isAdmin)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data ringkasan keuangan", nil, err.Error())
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil mengambil data ringkasan keuangan", nil, summary)
}
