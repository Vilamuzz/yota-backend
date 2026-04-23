package foster_children

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
	GetFosterChildrenList(ctx context.Context, params FosterChildrenQueryParams) pkg.Response
	GetFosterChildrenByID(ctx context.Context, id string) pkg.Response
	CreateFosterChildren(ctx context.Context, req CreateFosterChildrenRequest) pkg.Response
	UpdateFosterChildren(ctx context.Context, id string, req UpdateFosterChildrenRequest) pkg.Response
	DeleteFosterChildren(ctx context.Context, id string) pkg.Response

	GetFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateQueryParams) pkg.Response
	GetFosterChildrenCandidateByID(ctx context.Context, id string) pkg.Response
	CreateFosterChildrenCandidate(ctx context.Context, accountID string, req CreateFosterChildrenCandidateRequest) pkg.Response
	UpdateFosterChildrenCandidateStatus(ctx context.Context, id string, req UpdateFosterChildrenCandidateStatusRequest) pkg.Response
	CancelFosterChildrenCandidate(ctx context.Context, accountID string, id string) pkg.Response
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

func (s *service) GetFosterChildrenList(ctx context.Context, params FosterChildrenQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.Category != "" {
		options["category"] = params.Category
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	fosterChildren, err := s.repo.FindAllFosterChildren(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch foster children", nil, nil)
	}

	hasNext := len(fosterChildren) > params.Limit
	if hasNext {
		fosterChildren = fosterChildren[:params.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := params.PrevCursor != ""
	if hasNext && len(fosterChildren) > 0 {
		last := fosterChildren[len(fosterChildren)-1]
		nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
	}
	if hasPrev && len(fosterChildren) > 0 {
		first := fosterChildren[0]
		prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, ToFosterChildrenListResponse(fosterChildren, pagination))
}

func (s *service) GetFosterChildrenByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid foster children ID format"}, nil)
	}

	fosterChildren, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Foster children not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, fosterChildren.ToFosterChildrenResponse())
}

func (s *service) CreateFosterChildren(ctx context.Context, req CreateFosterChildrenRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Name == "" {
		errValidation["name"] = "Name is required"
	}
	if req.Gender == "" {
		errValidation["gender"] = "Gender is required"
	} else if req.Gender != Male && req.Gender != Female {
		errValidation["gender"] = "Invalid gender"
	}
	if req.Category == "" {
		errValidation["category"] = "Category is required"
	} else if req.Category != CategoryFatherless && req.Category != CategoryMotherless && req.Category != CategoryOrphan {
		errValidation["category"] = "Invalid category"
	}
	if req.BirthDate == "" {
		errValidation["birth_date"] = "Birth date is required"
	}
	if req.BirthPlace == "" {
		errValidation["birth_place"] = "Birth place is required"
	}
	if req.Address == "" {
		errValidation["address"] = "Address is required"
	}
	if req.ProfilePicture == nil {
		errValidation["profile_picture"] = "Profile picture is required"
	}
	if req.FamilyCard == nil {
		errValidation["family_card"] = "Family card is required"
	}
	if req.SKTM == nil {
		errValidation["sktm"] = "SKTM is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"birth_date": "Invalid date format, expected YYYY-MM-DD"}, nil)
	}

	// Upload profile picture
	profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload profile picture", nil, nil)
	}

	// Upload family card
	familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload family card", nil, nil)
	}

	// Upload SKTM
	sktmURL, err := s.s3Client.UploadFile(ctx, req.SKTM, "foster-children")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload SKTM", nil, nil)
	}

	now := time.Now()
	fosterChildrenID := uuid.New()

	fosterChildren := &FosterChildren{
		ID:             fosterChildrenID,
		Name:           req.Name,
		ProfilePicture: profilePictureURL,
		Gender:         req.Gender,
		IsGraduated:    req.IsGraduated,
		Category:       req.Category,
		BirthDate:      birthDate,
		BirthPlace:     req.BirthPlace,
		Address:        req.Address,
		FamilyCard:     familyCardURL,
		SKTM:           sktmURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Upload achievements
	if len(req.Achievements) > 0 {
		var achievements []Achivement
		for _, file := range req.Achievements {
			achievementURL, err := s.s3Client.UploadFile(ctx, file, "foster-children/achievements")
			if err != nil {
				logrus.WithError(err).Error("failed to upload achievement")
				continue
			}
			achievements = append(achievements, Achivement{
				ID:               uuid.New(),
				FosterChildrenID: fosterChildrenID,
				URL:              achievementURL,
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
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create foster children", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "CREATE", "foster_children", fosterChildren.ID.String(), nil, fosterChildren.ToFosterChildrenResponse())
	return pkg.NewResponse(http.StatusCreated, "Foster children successfully created", nil, fosterChildren.ToFosterChildrenResponse())
}

func (s *service) UpdateFosterChildren(ctx context.Context, id string, req UpdateFosterChildrenRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid foster children ID format"}, nil)
	}

	existing, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Foster children not found", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})

	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Gender != "" {
		if req.Gender != Male && req.Gender != Female {
			errValidation["gender"] = "Invalid gender"
		} else {
			updateData["gender"] = req.Gender
		}
	}
	if req.IsGraduated != nil {
		updateData["is_graduated"] = *req.IsGraduated
	}
	if req.Category != "" {
		if req.Category != CategoryFatherless && req.Category != CategoryMotherless && req.Category != CategoryOrphan {
			errValidation["category"] = "Invalid category"
		} else {
			updateData["category"] = req.Category
		}
	}
	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			errValidation["birth_date"] = "Invalid date format, expected YYYY-MM-DD"
		} else {
			updateData["birth_date"] = birthDate
		}
	}
	if req.BirthPlace != "" {
		updateData["birth_place"] = req.BirthPlace
	}
	if req.Address != "" {
		updateData["address"] = req.Address
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Upload profile picture
	if req.ProfilePicture != nil {
		existingImage := s3_pkg.ExtractObjectNameFromURL(existing.ProfilePicture)
		if err := s.s3Client.DeleteFile(ctx, existingImage); err != nil {
			logrus.WithError(err).Warn("failed to delete existing profile picture from S3")
		}
		profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload profile picture", nil, nil)
		}
		updateData["profile_picture"] = profilePictureURL
	}

	// Upload family card
	if req.FamilyCard != nil {
		existingImage := s3_pkg.ExtractObjectNameFromURL(existing.FamilyCard)
		if err := s.s3Client.DeleteFile(ctx, existingImage); err != nil {
			logrus.WithError(err).Warn("failed to delete existing family card from S3")
		}
		familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload family card", nil, nil)
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
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload SKTM", nil, nil)
		}
		updateData["sktm"] = sktmURL
	}

	// Handle achievements replacement
	if len(req.Achievements) > 0 {
		// Delete existing achievements from S3
		for _, ach := range existing.Achivements {
			objectName := s3_pkg.ExtractObjectNameFromURL(ach.URL)
			if err := s.s3Client.DeleteFile(ctx, objectName); err != nil {
				logrus.WithError(err).Warn("failed to delete existing achievement from S3")
			}
		}
		// Delete existing achievements from DB
		if err := s.repo.DeleteAchievementsByFosterChildrenID(ctx, id); err != nil {
			logrus.WithError(err).Error("failed to delete existing achievements")
		}

		// Upload new achievements
		var achievements []Achivement
		now := time.Now()
		for _, file := range req.Achievements {
			achievementURL, err := s.s3Client.UploadFile(ctx, file, "foster-children/achievements")
			if err != nil {
				logrus.WithError(err).Error("failed to upload achievement")
				continue
			}
			achievements = append(achievements, Achivement{
				ID:               uuid.New(),
				FosterChildrenID: existing.ID,
				URL:              achievementURL,
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

	if len(updateData) == 0 && len(req.Achievements) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	if len(updateData) > 0 {
		updateData["updated_at"] = time.Now()

		if err := s.repo.UpdateFosterChildren(ctx, id, updateData); err != nil {
			logrus.WithFields(logrus.Fields{
				"component":          "foster_children.service",
				"foster_children_id": id,
			}).WithError(err).Error("failed to update foster children")
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to update foster children", nil, nil)
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foster_children", id, existing.ToFosterChildrenResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Foster children updated successfully", nil, nil)
}

func (s *service) DeleteFosterChildren(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid foster children ID format"}, nil)
	}

	fosterChildren, err := s.repo.FindOneFosterChildren(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Foster children not found", nil, nil)
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
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to delete foster children", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "DELETE", "foster_children", id, fosterChildren.ToFosterChildrenResponse(), nil)
	return pkg.NewResponse(http.StatusOK, "Foster children deleted successfully", nil, nil)
}

func (s *service) GetFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.AccountID != "" {
		options["account_id"] = params.AccountID
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	candidates, err := s.repo.FindAllFosterChildrenCandidates(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch foster children candidates", nil, nil)
	}

	hasNext := len(candidates) > params.Limit
	if hasNext {
		candidates = candidates[:params.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := params.PrevCursor != ""
	if hasNext && len(candidates) > 0 {
		last := candidates[len(candidates)-1]
		nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
	}
	if hasPrev && len(candidates) > 0 {
		first := candidates[0]
		prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
	}

	pagination := pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, ToFosterChildrenCandidateListResponse(candidates, pagination))
}

func (s *service) GetFosterChildrenCandidateByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"id": "Invalid format"}, nil)
	}

	candidate, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Candidate not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, candidate.ToFosterChildrenCandidateResponse())
}

func (s *service) CreateFosterChildrenCandidate(ctx context.Context, accountID string, req CreateFosterChildrenCandidateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Name == "" {
		errValidation["name"] = "Name is required"
	}
	if req.Gender == "" {
		errValidation["gender"] = "Gender is required"
	}
	if req.Category == "" {
		errValidation["category"] = "Category is required"
	}
	if req.BirthDate == "" {
		errValidation["birth_date"] = "Birth date is required"
	}
	if req.BirthPlace == "" {
		errValidation["birth_place"] = "Birth place is required"
	}
	if req.Address == "" {
		errValidation["address"] = "Address is required"
	}
	if req.ProfilePicture == nil {
		errValidation["profile_picture"] = "Profile picture is required"
	}
	if req.FamilyCard == nil {
		errValidation["family_card"] = "Family card is required"
	}
	if req.SKTM == nil {
		errValidation["sktm"] = "SKTM is required"
	}
	if req.SubmitterName == "" {
		errValidation["submitter_name"] = "Submitter name is required"
	}
	if req.SubmitterPhone == "" {
		errValidation["submitter_phone"] = "Submitter phone is required"
	}
	if req.SubmitterAddress == "" {
		errValidation["submitter_address"] = "Submitter address is required"
	}
	if req.SubmitterIDCard == nil {
		errValidation["submitter_id_card"] = "Submitter ID card is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"birth_date": "Invalid date format, expected YYYY-MM-DD"}, nil)
	}

	profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload profile picture", nil, nil)
	}

	familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload family card", nil, nil)
	}

	sktmURL, err := s.s3Client.UploadFile(ctx, req.SKTM, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload SKTM", nil, nil)
	}

	submitterIDCardURL, err := s.s3Client.UploadFile(ctx, req.SubmitterIDCard, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload ID card", nil, nil)
	}

	now := time.Now()
	candidate := &FosterChildrenCandidate{
		ID:               uuid.New(),
		Name:             req.Name,
		ProfilePicture:   profilePictureURL,
		Gender:           req.Gender,
		Category:         req.Category,
		BirthDate:        birthDate,
		BirthPlace:       req.BirthPlace,
		Address:          req.Address,
		FamilyCard:       familyCardURL,
		SKTM:             sktmURL,
		SubmitterName:    req.SubmitterName,
		SubmitterPhone:   req.SubmitterPhone,
		SubmitterAddress: req.SubmitterAddress,
		SubmitterIDCard:  submitterIDCardURL,
		SubmittedBy:      uuid.MustParse(accountID),
		Status:           StatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateFosterChildrenCandidate(ctx, candidate); err != nil {
		logrus.WithError(err).Error("failed to create foster children candidate")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create candidate", nil, nil)
	}

	s.logService.CreateLog(ctx, &accountID, "CREATE", "foster_children_candidate", candidate.ID.String(), nil, candidate.ToFosterChildrenCandidateResponse())
	return pkg.NewResponse(http.StatusCreated, "Candidate created successfully", nil, candidate.ToFosterChildrenCandidateResponse())
}

func (s *service) UpdateFosterChildrenCandidateStatus(ctx context.Context, id string, req UpdateFosterChildrenCandidateStatusRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if req.Status != StatusAccepted && req.Status != StatusRejected {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"status": "Invalid status, must be accepted or rejected"}, nil)
	}

	existing, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Candidate not found", nil, nil)
	}

	if existing.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Only pending candidates can be updated", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	if req.Status == StatusRejected {
		updateData["rejection_reason"] = req.RejectionReason
	} else {
		updateData["rejection_reason"] = ""
	}

	if err := s.repo.UpdateFosterChildrenCandidate(ctx, id, updateData); err != nil {
		logrus.WithError(err).Error("failed to update candidate status")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update candidate status", nil, nil)
	}

	if req.Status == StatusAccepted {
		fc := &FosterChildren{
			ID:             uuid.New(),
			Name:           existing.Name,
			ProfilePicture: existing.ProfilePicture,
			Gender:         existing.Gender,
			IsGraduated:    false,
			Category:       existing.Category,
			BirthDate:      existing.BirthDate,
			BirthPlace:     existing.BirthPlace,
			Address:        existing.Address,
			FamilyCard:     existing.FamilyCard,
			SKTM:           existing.SKTM,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := s.repo.CreateFosterChildren(ctx, fc); err != nil {
			logrus.WithError(err).Error("failed to create foster children from accepted candidate")
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foster_children_candidate", id, existing.ToFosterChildrenCandidateResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Candidate status updated successfully", nil, nil)
}

func (s *service) CancelFosterChildrenCandidate(ctx context.Context, accountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	existing, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Candidate not found", nil, nil)
	}

	if existing.SubmittedBy.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "You are not authorized to cancel this candidate", nil, nil)
	}

	if existing.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Only pending candidates can be cancelled", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusCanceled,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateFosterChildrenCandidate(ctx, id, updateData); err != nil {
		logrus.WithError(err).Error("failed to cancel candidate")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to cancel candidate", nil, nil)
	}

	s.logService.CreateLog(ctx, &accountID, "UPDATE", "foster_children_candidate", id, existing.ToFosterChildrenCandidateResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Candidate cancelled successfully", nil, nil)
}
