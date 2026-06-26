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
	GetPrayerList(ctx context.Context, isAdmin bool, params PrayerQueryParams) pkg.Response
	GetPrayerByID(ctx context.Context, prayerID, accountID string) pkg.Response
	DeletePrayer(ctx context.Context, prayerID string) pkg.Response
	CreatePrayerAmen(ctx context.Context, prayerID, accountID string) pkg.Response
	CreateReportPrayer(ctx context.Context, prayerID, accountID string) pkg.Response
	AllowPrayer(ctx context.Context, prayerID string) pkg.Response
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
		return pkg.NewResponse(http.StatusOK, "Amin berhasil dihapus", nil, map[string]interface{}{
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
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menambahkan Amin", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Amin berhasil ditambahkan", nil, map[string]interface{}{
		"isAmen": true,
	})
}

func (s *service) CreateReportPrayer(ctx context.Context, prayerID, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}
	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}
	prayer, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	if prayer.Reported != nil && !*prayer.Reported {
		return pkg.NewResponse(http.StatusOK, "Doa berhasil dilaporkan", nil, nil)
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

func (s *service) AllowPrayer(ctx context.Context, prayerID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}

	prayer, err := s.repo.FindOnePrayer(ctx, map[string]interface{}{"id": prayerID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Doa tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}
	reported := false
	prayer.Reported = &reported
	if err := s.repo.UpdatePrayer(ctx, prayer); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "prayer.service",
			"prayer_id": prayerID,
		}).WithError(err).Error("failed to update prayer")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui doa", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Doa berhasil diizinkan", nil, nil)
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

func (s *service) GetPrayerList(ctx context.Context, isAdmin bool, params PrayerQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	options := map[string]interface{}{
		"limit":      params.Limit,
		"page":       params.Page,
		"account_id": params.AccountID,
	}

	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}

	if isAdmin {
		options["reported"] = true
	}

	if params.DonationSlug != "" {
		program, err := s.donationRepo.FindOneDonationProgram(ctx, map[string]interface{}{"slug": params.DonationSlug})
		if err != nil {
			emptyPagination := pkg.OffsetPagination{
				Page:       params.Page,
				Limit:      params.Limit,
				Total:      0,
				TotalPages: 0,
			}
			if isAdmin {
				return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toAdminPrayerListResponse([]Prayer{}, emptyPagination))
			}
			return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toPrayerListResponse([]Prayer{}, emptyPagination))
		}
		options["donation_program_id"] = program.ID.String()
	}

	total, err := s.repo.CountPrayers(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil jumlah data doa", nil, nil)
	}

	prayers, err := s.repo.FindAllPrayers(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan doa", nil, nil)
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit != 0 {
		totalPages++
	}

	pagination := pkg.OffsetPagination{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	if isAdmin {
		return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toAdminPrayerListResponse(prayers, pagination))
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan doa", nil, toPrayerListResponse(prayers, pagination))
}

func (s *service) DeletePrayer(ctx context.Context, prayerID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(prayerID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID doa tidak valid"}, nil)
	}

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
