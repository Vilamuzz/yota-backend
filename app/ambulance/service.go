package ambulance

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	CreateAmbulance(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response
	FindAmbulanceById(ctx context.Context, id string) pkg.Response
	ListAmbulance(ctx context.Context, queryParams AmbulanceQueryParams) pkg.Response
	UpdateAmbulance(ctx context.Context, id string, payload UpdateAmbulanceRequest) pkg.Response
	DeleteAmbulance(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo     Repository
	timeout  time.Duration
	s3Client s3_pkg.Client
}

func NewService(repo Repository, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{repo: repo, s3Client: s3Client, timeout: timeout}
}

func (s *service) CreateAmbulance(ctx context.Context, payload CreateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if err := uuid.Validate(payload.DriverID); err != nil {
		errValidation["driverId"] = "Format Driver tidak valid"
	}
	if payload.Image == nil {
		errValidation["image"] = "Gambar wajib diisi"
	}
	if payload.PlateNumber == "" {
		errValidation["plateNumber"] = "Nomor plat wajib diisi"
	}
	if payload.Status == "" {
		errValidation["status"] = "Status wajib diisi"
	} else {
		if payload.Status != AmbulanceStatusAvailable && payload.Status != AmbulanceStatusInUse && payload.Status != AmbulanceStatusMaintenance {
			errValidation["status"] = "Status tidak valid"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var imageURL string
	if payload.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.Image, "ambulances")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "ambulance.service",
				"plate":     payload.PlateNumber,
			}).WithError(err).Error("failed to upload ambulance image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar ambulans", nil, nil)
		}
		imageURL = uploadedURL
	}

	now := time.Now()
	ambulance := &Ambulance{
		ID:          uuid.New(),
		DriverID:    uuid.MustParse(payload.DriverID),
		PlateNumber: payload.PlateNumber,
		Image:       imageURL,
		Status:      payload.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateAmbulance(ctx, ambulance); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat data ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Ambulans berhasil dibuat", nil, nil)
}

func (s *service) FindAmbulanceById(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID ambulans tidak valid"}, nil)
	}
	ambulance, err := s.repo.FindOneAmbulance(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan data ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan data ambulans", nil, ambulance)
}

func (s *service) ListAmbulance(ctx context.Context, queryParams AmbulanceQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	ambulances, err := s.repo.FindAllAmbulances(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data ambulans", nil, nil)
	}
	hasNext := len(ambulances) > queryParams.Limit
	if hasNext {
		ambulances = ambulances[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulances) > 0 {
		lastAmbulance := ambulances[len(ambulances)-1]
		nextCursor = pkg.EncodeCursor(lastAmbulance.CreatedAt, lastAmbulance.ID.String())
	}
	if hasPrev && len(ambulances) > 0 {
		firstAmbulance := ambulances[0]
		prevCursor = pkg.EncodeCursor(firstAmbulance.CreatedAt, firstAmbulance.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toAmbulanceListResponse(ambulances, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) UpdateAmbulance(ctx context.Context, id string, payload UpdateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulance, err := s.repo.FindOneAmbulance(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan data ambulans", nil, nil)
	}

	updateData := make(map[string]interface{})
	if payload.PlateNumber != "" {
		updateData["plate_number"] = payload.PlateNumber
	}
	if payload.DriverID != "" {
		if err := uuid.Validate(payload.DriverID); err != nil {
			return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"driverId": "Format Driver tidak valid"}, nil)
		}
		updateData["driver_id"] = uuid.MustParse(payload.DriverID)
	}
	if payload.Status != "" {
		if payload.Status != AmbulanceStatusAvailable && payload.Status != AmbulanceStatusInUse && payload.Status != AmbulanceStatusMaintenance {
			return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"status": "Status tidak valid"}, nil)
		}
		updateData["status"] = payload.Status
	}
	if payload.Image != nil {
		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.Image, "ambulances")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "ambulance.service",
				"id":        id,
			}).WithError(err).Error("failed to upload new ambulance image")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar baru ambulans", nil, nil)
		}

		if ambulance.Image != "" {
			oldObjectName := s3_pkg.ExtractObjectNameFromURL(ambulance.Image)
			_ = s.s3Client.DeleteFile(ctx, oldObjectName)
		}
		updateData["image"] = uploadedURL
	}

	if err := s.repo.UpdateAmbulance(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui data ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Data ambulans berhasil diperbarui", nil, nil)
}

func (s *service) DeleteAmbulance(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.repo.DeleteAmbulance(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus data ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Data ambulans berhasil dihapus", nil, nil)
}
