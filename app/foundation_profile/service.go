package foundation_profile

import (
	"context"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetFoundationProfile(ctx context.Context) pkg.Response
	CreateFoundationProfile(ctx context.Context, payload FoundationProfileCreateRequest) pkg.Response
	UpdateFoundationProfile(ctx context.Context, id string, payload FoundationProfileUpdateRequest) pkg.Response
}

type service struct {
	repo       Repository
	logService app_log.Service
	s3Client   s3_pkg.Client
	timeout    time.Duration
}

func NewService(repo Repository, logService app_log.Service, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:       repo,
		logService: logService,
		s3Client:   s3Client,
		timeout:    timeout,
	}
}

func (s *service) GetFoundationProfile(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	profile, err := s.repo.FindFoundationProfile(ctx, map[string]interface{}{})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Profil yayasan tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, profile.toFoundationProfileResponse())
}

func (s *service) CreateFoundationProfile(ctx context.Context, payload FoundationProfileCreateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	if payload.FoundationName == "" {
		errValidation["foundation_name"] = "Nama yayasan wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// Check if a profile already exists (singleton pattern)
	existing, _ := s.repo.FindFoundationProfile(ctx, map[string]interface{}{})
	if existing != nil {
		errValidation["foundation_name"] = "Profil yayasan sudah ada. Hanya boleh ada 1 profil yayasan."
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	timeNow := time.Now()
	profile := &FoundationProfile{
		ID:                  uuid.New(),
		FoundationName:      payload.FoundationName,
		FounderName:         payload.FounderName,
		FoundationAddress:   payload.FoundationAddress,
		FoundationPhone:     payload.FoundationPhone,
		FoundationEmail:     payload.FoundationEmail,
		FoundationInstagram: payload.FoundationInstagram,
		FoundationFacebook:  payload.FoundationFacebook,
		FoundationTwitter:   payload.FoundationTwitter,
		EmbeddedAddress:     payload.EmbeddedAddress,
		CreatedAt:           timeNow,
		UpdatedAt:           timeNow,
	}

	if payload.FounderPicture != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.FounderPicture, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload founder picture")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah foto pendiri", nil, nil)
		}
		profile.FounderPicture = url
	}

	if payload.Logo != nil {
		url, err := s.s3Client.UploadFileOriginal(ctx, payload.Logo, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload logo")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah logo", nil, nil)
		}
		profile.Logo = url
	}

	if payload.Icon != nil {
		url, err := s.s3Client.UploadFileOriginal(ctx, payload.Icon, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload icon")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah ikon", nil, nil)
		}
		profile.Icon = url
	}

	if payload.OrganizationStructure != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.OrganizationStructure, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload organization structure")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah struktur organisasi", nil, nil)
		}
		profile.OrganizationStructure = url
	}

	if payload.HeroImageOne != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageOne, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image one")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 1", nil, nil)
		}
		profile.HeroImageOne = url
	}

	if payload.HeroImageTwo != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageTwo, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image two")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 2", nil, nil)
		}
		profile.HeroImageTwo = url
	}

	if payload.HeroImageThree != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageThree, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image three")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 3", nil, nil)
		}
		profile.HeroImageThree = url
	}

	if payload.HeroImageFour != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageFour, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image four")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 4", nil, nil)
		}
		profile.HeroImageFour = url
	}

	if err := s.repo.CreateFoundationProfile(ctx, profile); err != nil {
		logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to create foundation profile")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat profil yayasan", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "foundation_profile", profile.ID.String(), nil, profile.toFoundationProfileResponse())

	return pkg.NewResponse(http.StatusCreated, "Profil yayasan berhasil dibuat", nil, profile.toFoundationProfileResponse())
}

func (s *service) UpdateFoundationProfile(ctx context.Context, id string, payload FoundationProfileUpdateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID profil yayasan tidak valid"}, nil)
	}

	existing, err := s.repo.FindFoundationProfile(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Profil yayasan tidak ditemukan", nil, nil)
	}

	updateData := make(map[string]interface{})

	if payload.FoundationName != "" {
		updateData["foundation_name"] = payload.FoundationName
	}

	if payload.FounderName != "" {
		updateData["founder_name"] = payload.FounderName
	}

	if payload.FoundationAddress != "" {
		updateData["foundation_address"] = payload.FoundationAddress
	}

	if payload.FoundationPhone != "" {
		updateData["foundation_phone"] = payload.FoundationPhone
	}

	if payload.FoundationEmail != "" {
		updateData["foundation_email"] = payload.FoundationEmail
	}

	if payload.FoundationInstagram != nil {
		updateData["foundation_instagram"] = payload.FoundationInstagram
	}

	if payload.FoundationFacebook != nil {
		updateData["foundation_facebook"] = payload.FoundationFacebook
	}

	if payload.FoundationTwitter != nil {
		updateData["foundation_twitter"] = payload.FoundationTwitter
	}

	if payload.EmbeddedAddress != "" {
		updateData["embedded_address"] = payload.EmbeddedAddress
	}

	if payload.FounderPicture != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.FounderPicture, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload founder picture")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah foto pendiri", nil, nil)
		}
		if existing.FounderPicture != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.FounderPicture))
		}
		updateData["founder_picture"] = url
	}

	if payload.Logo != nil {
		url, err := s.s3Client.UploadFileOriginal(ctx, payload.Logo, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload logo")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah logo", nil, nil)
		}
		if existing.Logo != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.Logo))
		}
		updateData["logo"] = url
	}

	if payload.Icon != nil {
		url, err := s.s3Client.UploadFileOriginal(ctx, payload.Icon, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload icon")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah ikon", nil, nil)
		}
		if existing.Icon != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.Icon))
		}
		updateData["icon"] = url
	}

	if payload.OrganizationStructure != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.OrganizationStructure, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload organization structure")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah struktur organisasi", nil, nil)
		}
		if existing.OrganizationStructure != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.OrganizationStructure))
		}
		updateData["organization_structure"] = url
	}

	if payload.HeroImageOne != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageOne, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image one")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 1", nil, nil)
		}
		if existing.HeroImageOne != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.HeroImageOne))
		}
		updateData["hero_image_one"] = url
	}

	if payload.HeroImageTwo != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageTwo, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image two")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 2", nil, nil)
		}
		if existing.HeroImageTwo != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.HeroImageTwo))
		}
		updateData["hero_image_two"] = url
	}

	if payload.HeroImageThree != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageThree, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image three")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 3", nil, nil)
		}
		if existing.HeroImageThree != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.HeroImageThree))
		}
		updateData["hero_image_three"] = url
	}

	if payload.HeroImageFour != nil {
		url, err := s.s3Client.UploadFile(ctx, payload.HeroImageFour, "foundation-profile")
		if err != nil {
			logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image four")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah hero image 4", nil, nil)
		}
		if existing.HeroImageFour != "" {
			_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(existing.HeroImageFour))
		}
		updateData["hero_image_four"] = url
	}

	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"updateData": "Tidak ada data untuk diperbarui"}, nil)
	}

	updateData["updated_at"] = time.Now()

	if err := s.repo.UpdateFoundationProfile(ctx, id, updateData); err != nil {
		logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to update foundation profile")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui profil yayasan", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foundation_profile", id, existing.toFoundationProfileResponse(), updateData)

	return pkg.NewResponse(http.StatusOK, "Profil yayasan berhasil diperbarui", nil, nil)
}
