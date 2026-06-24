package foundation_profile

import (
	"context"
	"errors"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

type uploadError struct {
	field string
	msg   string
	err   error
}

func (e uploadError) Error() string {
	return e.msg
}

func (s *service) CreateFoundationProfile(ctx context.Context, payload FoundationProfileCreateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)

	if payload.FoundationName == "" {
		errValidation["foundation_name"] = "Nama yayasan wajib diisi"
	}

	if payload.FoundationPhone != "" && !pkg.IsValidPhoneNumber(payload.FoundationPhone) {
		errValidation["foundation_phone"] = "Format nomor telepon tidak valid"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

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

	var (
		founderPictureURL        string
		logoURL                  string
		iconURL                  string
		organizationStructureURL string
		heroImageOneURL          string
		heroImageTwoURL          string
		heroImageThreeURL        string
		heroImageFourURL         string
	)

	g, gCtx := errgroup.WithContext(ctx)

	if payload.FounderPicture != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.FounderPicture, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload founder picture")
				return uploadError{field: "founder_picture", msg: "Gagal mengunggah foto pendiri", err: err}
			}
			founderPictureURL = url
			return nil
		})
	}

	if payload.Logo != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFileOriginal(gCtx, payload.Logo, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload logo")
				return uploadError{field: "logo", msg: "Gagal mengunggah logo", err: err}
			}
			logoURL = url
			return nil
		})
	}

	if payload.Icon != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFileOriginal(gCtx, payload.Icon, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload icon")
				return uploadError{field: "icon", msg: "Gagal mengunggah ikon", err: err}
			}
			iconURL = url
			return nil
		})
	}

	if payload.OrganizationStructure != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.OrganizationStructure, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload organization structure")
				return uploadError{field: "organization_structure", msg: "Gagal mengunggah struktur organisasi", err: err}
			}
			organizationStructureURL = url
			return nil
		})
	}

	if payload.HeroImageOne != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageOne, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image one")
				return uploadError{field: "hero_image_one", msg: "Gagal mengunggah hero image 1", err: err}
			}
			heroImageOneURL = url
			return nil
		})
	}

	if payload.HeroImageTwo != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageTwo, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image two")
				return uploadError{field: "hero_image_two", msg: "Gagal mengunggah hero image 2", err: err}
			}
			heroImageTwoURL = url
			return nil
		})
	}

	if payload.HeroImageThree != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageThree, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image three")
				return uploadError{field: "hero_image_three", msg: "Gagal mengunggah hero image 3", err: err}
			}
			heroImageThreeURL = url
			return nil
		})
	}

	if payload.HeroImageFour != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageFour, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image four")
				return uploadError{field: "hero_image_four", msg: "Gagal mengunggah hero image 4", err: err}
			}
			heroImageFourURL = url
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		var uErr uploadError
		if errors.As(err, &uErr) {
			return pkg.NewResponse(http.StatusInternalServerError, uErr.msg, nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file", nil, nil)
	}

	if founderPictureURL != "" {
		profile.FounderPicture = founderPictureURL
	}
	if logoURL != "" {
		profile.Logo = logoURL
	}
	if iconURL != "" {
		profile.Icon = iconURL
	}
	if organizationStructureURL != "" {
		profile.OrganizationStructure = organizationStructureURL
	}
	if heroImageOneURL != "" {
		profile.HeroImageOne = heroImageOneURL
	}
	if heroImageTwoURL != "" {
		profile.HeroImageTwo = heroImageTwoURL
	}
	if heroImageThreeURL != "" {
		profile.HeroImageThree = heroImageThreeURL
	}
	if heroImageFourURL != "" {
		profile.HeroImageFour = heroImageFourURL
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
		if !pkg.IsValidPhoneNumber(payload.FoundationPhone) {
			return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"foundation_phone": "Format nomor telepon tidak valid"}, nil)
		}
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

	var (
		founderPictureURL        string
		logoURL                  string
		iconURL                  string
		organizationStructureURL string
		heroImageOneURL          string
		heroImageTwoURL          string
		heroImageThreeURL        string
		heroImageFourURL         string
	)

	g, gCtx := errgroup.WithContext(ctx)

	if payload.FounderPicture != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.FounderPicture, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload founder picture")
				return uploadError{field: "founder_picture", msg: "Gagal mengunggah foto pendiri", err: err}
			}
			founderPictureURL = url
			return nil
		})
	}

	if payload.Logo != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFileOriginal(gCtx, payload.Logo, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload logo")
				return uploadError{field: "logo", msg: "Gagal mengunggah logo", err: err}
			}
			logoURL = url
			return nil
		})
	}

	if payload.Icon != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFileOriginal(gCtx, payload.Icon, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload icon")
				return uploadError{field: "icon", msg: "Gagal mengunggah ikon", err: err}
			}
			iconURL = url
			return nil
		})
	}

	if payload.OrganizationStructure != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.OrganizationStructure, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload organization structure")
				return uploadError{field: "organization_structure", msg: "Gagal mengunggah struktur organisasi", err: err}
			}
			organizationStructureURL = url
			return nil
		})
	}

	if payload.HeroImageOne != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageOne, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image one")
				return uploadError{field: "hero_image_one", msg: "Gagal mengunggah hero image 1", err: err}
			}
			heroImageOneURL = url
			return nil
		})
	}

	if payload.HeroImageTwo != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageTwo, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image two")
				return uploadError{field: "hero_image_two", msg: "Gagal mengunggah hero image 2", err: err}
			}
			heroImageTwoURL = url
			return nil
		})
	}

	if payload.HeroImageThree != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageThree, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image three")
				return uploadError{field: "hero_image_three", msg: "Gagal mengunggah hero image 3", err: err}
			}
			heroImageThreeURL = url
			return nil
		})
	}

	if payload.HeroImageFour != nil {
		g.Go(func() error {
			url, err := s.s3Client.UploadFile(gCtx, payload.HeroImageFour, "foundation-profile")
			if err != nil {
				logrus.WithField("component", "foundation_profile.service").WithError(err).Error("failed to upload hero image four")
				return uploadError{field: "hero_image_four", msg: "Gagal mengunggah hero image 4", err: err}
			}
			heroImageFourURL = url
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		var uErr uploadError
		if errors.As(err, &uErr) {
			return pkg.NewResponse(http.StatusInternalServerError, uErr.msg, nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file", nil, nil)
	}

	var filesToDelete []string

	if payload.FounderPicture != nil {
		if existing.FounderPicture != "" {
			filesToDelete = append(filesToDelete, existing.FounderPicture)
		}
		updateData["founder_picture"] = founderPictureURL
	}

	if payload.Logo != nil {
		if existing.Logo != "" {
			filesToDelete = append(filesToDelete, existing.Logo)
		}
		updateData["logo"] = logoURL
	}

	if payload.Icon != nil {
		if existing.Icon != "" {
			filesToDelete = append(filesToDelete, existing.Icon)
		}
		updateData["icon"] = iconURL
	}

	if payload.OrganizationStructure != nil {
		if existing.OrganizationStructure != "" {
			filesToDelete = append(filesToDelete, existing.OrganizationStructure)
		}
		updateData["organization_structure"] = organizationStructureURL
	}

	if payload.HeroImageOne != nil {
		if existing.HeroImageOne != "" {
			filesToDelete = append(filesToDelete, existing.HeroImageOne)
		}
		updateData["hero_image_one"] = heroImageOneURL
	}

	if payload.HeroImageTwo != nil {
		if existing.HeroImageTwo != "" {
			filesToDelete = append(filesToDelete, existing.HeroImageTwo)
		}
		updateData["hero_image_two"] = heroImageTwoURL
	}

	if payload.HeroImageThree != nil {
		if existing.HeroImageThree != "" {
			filesToDelete = append(filesToDelete, existing.HeroImageThree)
		}
		updateData["hero_image_three"] = heroImageThreeURL
	}

	if payload.HeroImageFour != nil {
		if existing.HeroImageFour != "" {
			filesToDelete = append(filesToDelete, existing.HeroImageFour)
		}
		updateData["hero_image_four"] = heroImageFourURL
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

	// Clean up old files after successful DB update
	for _, oldFile := range filesToDelete {
		_ = s.s3Client.DeleteFile(ctx, s3_pkg.ExtractObjectNameFromURL(oldFile))
	}

	return pkg.NewResponse(http.StatusOK, "Profil yayasan berhasil diperbarui", nil, nil)
}
