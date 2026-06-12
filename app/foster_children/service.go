package foster_children

import (
	"context"
	"fmt"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	GetFosterChildrenList(ctx context.Context, params FosterChildrenQueryParams, isAdmin bool) pkg.Response
	GetFosterChildrenByID(ctx context.Context, id string) pkg.Response
	GetFosterChildrenBySlug(ctx context.Context, slug string) pkg.Response
	CreateFosterChildren(ctx context.Context, req CreateFosterChildrenRequest) pkg.Response
	UpdateFosterChildren(ctx context.Context, id string, req UpdateFosterChildrenRequest) pkg.Response
	DeleteFosterChildren(ctx context.Context, id string) pkg.Response
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

func (s *service) GetFosterChildrenList(ctx context.Context, params FosterChildrenQueryParams, isAdmin bool) pkg.Response {
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
		"limit": params.Limit,
		"page":  params.Page,
	}
	if isAdmin {
		options["is_admin"] = true
	}
	if params.Category != "" {
		options["category"] = params.Category
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.Gender != "" {
		options["gender"] = params.Gender
	}
	if params.IsGraduated != nil {
		options["is_graduated"] = *params.IsGraduated
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}

	total, err := s.repo.CountFosterChildren(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data anak asuh", nil, nil)
	}

	fosterChildren, err := s.repo.FindAllFosterChildren(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data anak asuh", nil, nil)
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
		return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ToAdminFosterChildrenListResponse(fosterChildren, pagination))
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ToFosterChildrenListResponse(fosterChildren, pagination))
}

func (s *service) GetFosterChildrenByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID anak asuh tidak valid"}, nil)
	}

	fosterChildren, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Anak asuh tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, fosterChildren.ToAdminFosterChildrenDetailResponse())
}

func (s *service) GetFosterChildrenBySlug(ctx context.Context, slug string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	fosterChildren, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"slug": slug})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Anak asuh tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, fosterChildren.ToFosterChildrenDetailResponse())
}

func (s *service) CreateFosterChildren(ctx context.Context, req CreateFosterChildrenRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Name == "" {
		errValidation["name"] = "Nama wajib diisi"
	}
	if req.Gender == "" {
		errValidation["gender"] = "Jenis kelamin wajib diisi"
	} else if req.Gender != Male && req.Gender != Female {
		errValidation["gender"] = "Jenis kelamin tidak valid"
	}
	if req.Category == "" {
		errValidation["category"] = "Kategori wajib diisi"
	} else if req.Category != CategoryFatherless && req.Category != CategoryMotherless && req.Category != CategoryOrphan {
		errValidation["category"] = "Kategori tidak valid"
	}
	if req.BirthDate == "" {
		errValidation["birthDate"] = "Tanggal lahir wajib diisi"
	}
	if req.BirthPlace == "" {
		errValidation["birthPlace"] = "Tempat lahir wajib diisi"
	}
	if req.SchoolName == "" {
		errValidation["schoolName"] = "Nama sekolah wajib diisi"
	}
	if req.EducationLevel <= 0 || req.EducationLevel > 12 {
		errValidation["educationLevel"] = "Tingkat pendidikan tidak valid (maksimal kelas 12)"
	}
	if req.Address == "" {
		errValidation["address"] = "Alamat wajib diisi"
	}
	if req.ProfilePicture == nil {
		errValidation["profilePicture"] = "Foto profil wajib diisi"
	}
	if req.FamilyCard == nil {
		errValidation["familyCard"] = "Kartu keluarga wajib diisi"
	}
	if req.SKTM == nil {
		errValidation["sktm"] = "SKTM wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"birthDate": "Format tanggal tidak valid, diharapkan YYYY-MM-DD"}, nil)
	}

	// Upload foto profil
	profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah foto profil", nil, nil)
	}

	// Upload kartu keluarga
	familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah kartu keluarga", nil, nil)
	}

	// Upload SKTM
	sktmURL, err := s.s3Client.UploadFile(ctx, req.SKTM, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah SKTM", nil, nil)
	}

	now := time.Now()
	fosterChildrenID := uuid.New()
	slug := fmt.Sprintf("%s-%s", pkg.Slugify(req.Name), fosterChildrenID.String()[:5])

	fosterChildren := &FosterChildren{
		ID:             fosterChildrenID,
		Slug:           slug,
		Name:           req.Name,
		ProfilePicture: profilePictureURL,
		Gender:         req.Gender,
		IsGraduated:    req.IsGraduated,
		Category:       req.Category,
		BirthDate:      birthDate,
		BirthPlace:     req.BirthPlace,
		SchoolName:     req.SchoolName,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		FamilyCard:     familyCardURL,
		SKTM:           sktmURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Upload piagam prestasi jika ada
	if len(req.Achievements) > 0 {
		var achievements []Achivement
		for i, file := range req.Achievements {
			achievementURL, err := s.s3Client.UploadFile(ctx, file, "foster-children/achievements")
			if err != nil {
				logrus.WithError(err).Error("failed to upload achievement")
				continue
			}

			note := ""
			if i < len(req.AchivementNotes) {
				note = req.AchivementNotes[i]
			}

			achievements = append(achievements, Achivement{
				ID:               uuid.New(),
				FosterChildrenID: fosterChildrenID,
				URL:              achievementURL,
				Note:             note,
				CreatedAt:        now,
				UpdatedAt:        now,
			})
		}
		fosterChildren.Achivements = achievements
	}

	if err := s.repo.CreateFosterChildren(ctx, fosterChildren); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "foster_children.service",
			"name":      req.Name,
		}).WithError(err).Error("failed to create foster children")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat data anak asuh", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "foster_children", fosterChildren.ID.String(), nil, fosterChildren.ToAdminFosterChildrenDetailResponse())
	return pkg.NewResponse(http.StatusCreated, "Anak asuh berhasil dibuat", nil, nil)
}

func (s *service) UpdateFosterChildren(ctx context.Context, id string, req UpdateFosterChildrenRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID anak asuh tidak valid"}, nil)
	}

	existing, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Anak asuh tidak ditemukan", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Gender != "" {
		if req.Gender != Male && req.Gender != Female {
			errValidation["gender"] = "Jenis kelamin tidak valid"
		} else {
			updateData["gender"] = req.Gender
		}
	}
	if req.IsGraduated != nil {
		updateData["is_graduated"] = *req.IsGraduated
	}
	if req.Category != "" {
		if req.Category != CategoryFatherless && req.Category != CategoryMotherless && req.Category != CategoryOrphan {
			errValidation["category"] = "Kategori tidak valid"
		} else {
			updateData["category"] = req.Category
		}
	}
	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			errValidation["birthDate"] = "Format tanggal tidak valid, diharapkan YYYY-MM-DD"
		} else {
			updateData["birthDate"] = birthDate
		}
	}
	if req.BirthPlace != "" {
		updateData["birth_place"] = req.BirthPlace
	}
	if req.SchoolName != "" {
		updateData["school_name"] = req.SchoolName
	}
	if req.EducationLevel != 0 {
		if req.EducationLevel < 1 || req.EducationLevel > 12 {
			errValidation["educationLevel"] = "Tingkat pendidikan tidak valid (maksimal kelas 12)"
		} else {
			updateData["education_level"] = req.EducationLevel
		}
	}
	if req.Address != "" {
		updateData["address"] = req.Address
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// Upload foto profil
	if req.ProfilePicture != nil {
		existingImage := s3_pkg.ExtractObjectNameFromURL(existing.ProfilePicture)
		if err := s.s3Client.DeleteFile(ctx, existingImage); err != nil {
			logrus.WithError(err).Warn("failed to delete existing profile picture from S3")
		}
		profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah foto profil", nil, nil)
		}
		updateData["profile_picture"] = profilePictureURL
	}

	// Upload kartu keluarga
	if req.FamilyCard != nil {
		existingImage := s3_pkg.ExtractObjectNameFromURL(existing.FamilyCard)
		if err := s.s3Client.DeleteFile(ctx, existingImage); err != nil {
			logrus.WithError(err).Warn("failed to delete existing family card from S3")
		}
		familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah kartu keluarga", nil, nil)
		}
		updateData["family_card"] = familyCardURL
	}

	// Upload SKTM
	if req.SKTM != nil {
		existingImage := s3_pkg.ExtractObjectNameFromURL(existing.SKTM)
		if err := s.s3Client.DeleteFile(ctx, existingImage); err != nil {
			logrus.WithError(err).Warn("failed to delete existing SKTM from S3")
		}
		sktmURL, err := s.s3Client.UploadFile(ctx, req.SKTM, "foster-children")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah SKTM", nil, nil)
		}
		updateData["sktm"] = sktmURL
	}

	// Handle achievements replacement
	if len(req.UpdateAchivementNotes) > 0 && len(req.UpdateAchivementNotes) != len(req.AchivementIDs) {
		errValidation["updateAchivementNotes"] = "Jumlah catatan prestasi yang diperbarui harus sama dengan jumlah ID prestasi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	// 1. Identify and Delete Missing Achievements
	payloadAchivementIDMap := make(map[string]bool)
	for _, aid := range req.AchivementIDs {
		payloadAchivementIDMap[aid] = true
	}

	for _, ach := range existing.Achivements {
		if !payloadAchivementIDMap[ach.ID.String()] {
			objectName := s3_pkg.ExtractObjectNameFromURL(ach.URL)
			if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
				logrus.WithError(err).Warn("failed to delete existing achievement from S3")
			}

			if err := s.repo.DeleteAchievementByID(ctx, ach.ID.String()); err != nil {
				logrus.WithError(err).Error("failed to delete existing achievement from DB")
			}
		}
	}

	// 2. Update Existing Achievements Metadata
	for i, aid := range req.AchivementIDs {
		updateAchivementData := make(map[string]interface{})
		if i < len(req.UpdateAchivementNotes) {
			updateAchivementData["note"] = req.UpdateAchivementNotes[i]
		}

		if len(updateAchivementData) > 0 {
			if err := s.repo.UpdateAchievement(ctx, aid, updateAchivementData); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":     "foster_children.service",
					"achivement_id": aid,
				}).WithError(err).Warn("failed to update achievement metadata")
			}
		}
	}

	// 3. Add New Achievements
	if len(req.Achievements) > 0 {
		var achievements []Achivement
		now := time.Now()
		for i, file := range req.Achievements {
			achievementURL, err := s.s3Client.UploadFile(ctx, file, "foster-children/achievements")
			if err != nil {
				logrus.WithError(err).Error("failed to upload achievement")
				continue
			}

			note := ""
			if i < len(req.AchivementNotes) {
				note = req.AchivementNotes[i]
			}

			achievements = append(achievements, Achivement{
				ID:               uuid.New(),
				FosterChildrenID: existing.ID,
				URL:              achievementURL,
				Note:             note,
				CreatedAt:        now,
				UpdatedAt:        now,
			})
		}
		if len(achievements) > 0 {
			if err := s.repo.CreateAchievements(ctx, achievements); err != nil {
				logrus.WithError(err).Error("failed to create achievements")
			}
		}
	}

	if len(updateData) == 0 && len(req.Achievements) == 0 && len(req.AchivementIDs) == 0 && len(existing.Achivements) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"updateData": "Tidak ada data untuk diperbarui"}, nil)
	}

	if len(updateData) > 0 {
		updateData["updated_at"] = time.Now()

		if err := s.repo.UpdateFosterChildren(ctx, id, updateData); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":          "foster_children.service",
				"foster_children_id": id,
			}).WithError(err).Error("failed to update foster children")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui data anak asuh", nil, nil)
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foster_children", id, existing.ToAdminFosterChildrenDetailResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Anak asuh berhasil diperbarui", nil, nil)
}

func (s *service) DeleteFosterChildren(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID anak asuh tidak valid"}, nil)
	}

	fosterChildren, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Anak asuh tidak ditemukan", nil, nil)
	}

	// Delete achievements from S3
	for _, ach := range fosterChildren.Achivements {
		objectName := s3_pkg.ExtractObjectNameFromURL(ach.URL)
		if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
			logrus.WithError(err).Warn("failed to delete achievement from S3")
		}
	}

	// Delete achievements from DB
	if err := s.repo.DeleteAchievementsByFosterChildrenID(ctx, id); err != nil {
		logrus.WithError(err).Error("failed to delete achievements")
	}

	// Delete S3 files
	if fosterChildren.ProfilePicture != "" {
		objectName := s3_pkg.ExtractObjectNameFromURL(fosterChildren.ProfilePicture)
		if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
			logrus.WithError(err).Warn("failed to delete profile picture from S3")
		}
	}
	if fosterChildren.FamilyCard != "" {
		objectName := s3_pkg.ExtractObjectNameFromURL(fosterChildren.FamilyCard)
		if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
			logrus.WithError(err).Warn("failed to delete family card from S3")
		}
	}
	if fosterChildren.SKTM != "" {
		objectName := s3_pkg.ExtractObjectNameFromURL(fosterChildren.SKTM)
		if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
			logrus.WithError(err).Warn("failed to delete SKTM from S3")
		}
	}

	if err := s.repo.DeleteFosterChildren(ctx, id); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":          "foster_children.service",
			"foster_children_id": id,
		}).WithError(err).Error("failed to delete foster children")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus data anak asuh", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "foster_children", id, fosterChildren.ToAdminFosterChildrenDetailResponse(), nil)
	return pkg.NewResponse(http.StatusOK, "Anak asuh berhasil dihapus", nil, nil)
}
