package prayer

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetPrayerList(ctx context.Context, accountID, donationSlug string, isAdmin bool, params PrayerQueryParams) pkg.Response
	GetPrayerByID(ctx context.Context, prayerID, accountID string) pkg.Response
	DeletePrayer(ctx context.Context, prayerID string) pkg.Response
	CreatePrayerAmen(ctx context.Context, prayerID, accountID string) pkg.Response
	CreateReportPrayer(ctx context.Context, prayerID, accountID string, payload ReportPrayerRequest) pkg.Response
}

type service struct {
	repo         Repository
	donationRepo donation_program.Repository
	timeout      time.Duration
}

func NewService(repo Repository, donationRepo donation_program.Repository, timeout time.Duration) Service {
	return &service{repo: repo, donationRepo: donationRepo, timeout: timeout}
}

func (s *service) CreatePrayerAmen(ctx context.Context, prayerID, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}
	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}

	_, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	rowsAffected, err := s.repo.DeleteAmen(ctx, prayerID, accountID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "prayer.service",
			"prayer_id":  prayerID,
			"account_id": accountID,
		}).WithError(err).Error("failed to delete amen")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete amen", nil, nil)
	}

	if rowsAffected > 0 {
		return pkg.NewResponse(http.StatusOK, "Aamiin berhasil dihapus", nil, map[string]interface{}{
			"isAmen": false,
		})
	}

	amen := &PrayerAmen{
		PrayerID:  uuid.MustParse(prayerID),
		AccountID: uuid.MustParse(accountID),
	}
	if err := s.repo.CreateAmen(ctx, amen); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "prayer.service",
			"prayer_id":  prayerID,
			"account_id": accountID,
		}).WithError(err).Error("failed to create amen")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menambahkan Aamiin", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Aamiin berhasil ditambahkan", nil, map[string]interface{}{
		"isAmen": true,
	})
}

func (s *service) CreateReportPrayer(ctx context.Context, prayerID, accountID string, payload ReportPrayerRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}
	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}
	_, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	if payload.Reason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan Validasi", map[string]string{
			"reason": "Alasan wajib diisi",
		}, nil)
	}

	_, err = s.repo.FindReport(ctx, map[string]interface{}{
		"prayer_id":  prayerID,
		"account_id": accountID,
	})
	if err == nil {
		return pkg.NewResponse(http.StatusOK, "Doa berhasil dilaporkan", nil, nil)
	}

	report := &PrayerReport{
		PrayerID:  uuid.MustParse(prayerID),
		AccountID: uuid.MustParse(accountID),
		Reason:    payload.Reason,
	}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "prayer.service",
			"prayer_id":  prayerID,
			"account_id": accountID,
		}).WithError(err).Error("failed to create prayer report")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal melaporkan doa", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Doa berhasil dilaporkan", nil, nil)
}

func (s *service) GetPrayerByID(ctx context.Context, prayerID, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}
	prayer, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID, "account_id": accountID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, prayer.toPrayerResponse())
}

func (s *service) GetPrayerList(ctx context.Context, accountID, donationSlug string, isAdmin bool, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	options := map[string]interface{}{
		"limit":      params.Limit,
		"account_id": accountID,
	}

	if isAdmin {
		options["reported"] = true
	}

	pagination := pkg.CursorPagination{
		Limit: params.Limit,
	}

	if donationSlug != "" {
		program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": donationSlug})
		if err != nil {
			return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toPrayerListResponse([]Prayer{}, pagination))
		}
		options["donation_program_id"] = program.ID.String()
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	prayers, err := s.repo.FindAllPrayers(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(prayers) > params.Limit
		hasNext = true
		if len(prayers) > params.Limit {
			prayers = prayers[:params.Limit]
		}
		for i, j := 0, len(prayers)-1; i < j; i, j = i+1, j-1 {
			prayers[i], prayers[j] = prayers[j], prayers[i]
		}
	} else {
		hasNext = len(prayers) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			prayers = prayers[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
	if len(prayers) > 0 {
		first := prayers[0]
		last := prayers[len(prayers)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	pagination = pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toPrayerListResponse(prayers, pagination))
}

func (s *service) DeletePrayer(ctx context.Context, prayerID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	if err := s.repo.DeletePrayer(ctx, prayerID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus doa", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Doa berhasil dihapus", nil, nil)
}
