package foster_children_candidate

import (
	"context"
	"net/http"
	"time"

	app_log "github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FosterChildrenCreator is a minimal interface to create a foster children record
// after a candidate is fully accepted, avoiding a circular import.
type FosterChildrenCreator interface {
	CreateFosterChildrenFromCandidate(ctx context.Context, candidateID, name, profilePicture, familyCard, sktm, birthPlace, schoolName, address string, gender string, category string, educationLevel int, birthDate time.Time) error
}

type Service interface {
	GetFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateAdminQueryParams) pkg.Response
	GetMyFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateQueryParams) pkg.Response
	GetFosterChildrenCandidateByID(ctx context.Context, id string) pkg.Response
	GetMyFosterChildrenCandidateByID(ctx context.Context, accountID string, id string) pkg.Response
	CreateFosterChildrenCandidate(ctx context.Context, accountID string, req CreateFosterChildrenCandidateRequest) pkg.Response
	AcceptFosterChildrenCandidate(ctx context.Context, id string, role enum.RoleName) pkg.Response
	RejectFosterChildrenCandidate(ctx context.Context, id string, req RejectFosterChildrenCandidateRequest) pkg.Response
	CancelFosterChildrenCandidate(ctx context.Context, accountID string, id string) pkg.Response
}

type service struct {
	repo               Repository
	fosterChildrenRepo FosterChildrenCreator
	logService         app_log.Service
	s3Client           s3_pkg.Client
	timeout            time.Duration
}

func NewService(repo Repository, fosterChildrenRepo FosterChildrenCreator, logService app_log.Service, s3Client s3_pkg.Client, timeout time.Duration) Service {
	return &service{
		repo:               repo,
		fosterChildrenRepo: fosterChildrenRepo,
		logService:         logService,
		s3Client:           s3Client,
		timeout:            timeout,
	}
}

func (s *service) GetFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateAdminQueryParams) pkg.Response {
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
	if params.Status != "" {
		options["status"] = params.Status
	}
	if params.AccountID != "" {
		options["account_id"] = params.AccountID
	}
	if params.Category != "" {
		options["category"] = params.Category
	}
	if params.Gender != "" {
		options["gender"] = params.Gender
	}
	if params.SortBy != "" {
		options["sort_by"] = params.SortBy
	}
	if params.Search != "" {
		options["search"] = params.Search
	}

	total, err := s.repo.CountFosterChildrenCandidates(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data calon anak asuh", nil, nil)
	}

	candidates, err := s.repo.FindAllFosterChildrenCandidates(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data calon anak asuh", nil, nil)
	}

	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := pkg.OffsetPagination{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ToFosterChildrenCandidateAdminListResponse(candidates, pagination))
}

func (s *service) GetMyFosterChildrenCandidateList(ctx context.Context, params FosterChildrenCandidateQueryParams) pkg.Response {
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
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data calon anak asuh", nil, nil)
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

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ToFosterChildrenCandidateListResponse(candidates, pagination))
}

func (s *service) GetFosterChildrenCandidateByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format tidak valid"}, nil)
	}

	candidate, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Calon tidak ditemukan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, candidate.ToFosterChildrenCandidateResponse())
}

func (s *service) GetMyFosterChildrenCandidateByID(ctx context.Context, accountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format tidak valid"}, nil)
	}

	candidate, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Calon tidak ditemukan", nil, nil)
	}

	if candidate.SubmittedBy.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melihat calon ini", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, candidate.ToFosterChildrenCandidateResponse())
}

func (s *service) CreateFosterChildrenCandidate(ctx context.Context, accountID string, req CreateFosterChildrenCandidateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Name == "" {
		errValidation["name"] = "Nama wajib diisi"
	}
	if req.Gender == "" {
		errValidation["gender"] = "Jenis kelamin wajib diisi"
	}
	if req.Category == "" {
		errValidation["category"] = "Kategori wajib diisi"
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
	if req.SubmitterName == "" {
		errValidation["submitterName"] = "Nama pengirim wajib diisi"
	}
	if req.SubmitterPhone == "" {
		errValidation["submitterPhone"] = "Nomor telepon pengirim wajib diisi"
	}
	if req.SubmitterAddress == "" {
		errValidation["submitterAddress"] = "Alamat pengirim wajib diisi"
	}
	if req.SubmitterIDCard == nil {
		errValidation["submitterIdCard"] = "KTP pengirim wajib diisi"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"birthDate": "Format tanggal tidak valid, diharapkan YYYY-MM-DD"}, nil)
	}

	profilePictureURL, err := s.s3Client.UploadFile(ctx, req.ProfilePicture, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah foto profil", nil, nil)
	}

	familyCardURL, err := s.s3Client.UploadFile(ctx, req.FamilyCard, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah kartu keluarga", nil, nil)
	}

	sktmURL, err := s.s3Client.UploadFile(ctx, req.SKTM, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah SKTM", nil, nil)
	}

	submitterIDCardURL, err := s.s3Client.UploadFile(ctx, req.SubmitterIDCard, "foster-children-candidates")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah KTP", nil, nil)
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
		SchoolName:       req.SchoolName,
		EducationLevel:   req.EducationLevel,
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
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat calon anak asuh", nil, nil)
	}

	s.logService.CreateLog(ctx, &accountID, "CREATE", "foster_children_candidate", candidate.ID.String(), nil, candidate.ToFosterChildrenCandidateResponse())
	return pkg.NewResponse(http.StatusCreated, "Calon anak asuh berhasil dibuat", nil, candidate.ToFosterChildrenCandidateResponse())
}

func (s *service) AcceptFosterChildrenCandidate(ctx context.Context, id string, role enum.RoleName) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	existing, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Calon tidak ditemukan", nil, nil)
	}

	var nextStatus Status
	var message string

	switch role {
	case enum.RoleSocialManager:
		if existing.Status != StatusPending {
			return pkg.NewResponse(http.StatusBadRequest, "Hanya calon dengan status pending yang dapat disetujui oleh Koordinator Sosial", nil, nil)
		}
		nextStatus = StatusSocialManagerAccepted
		message = "Calon berhasil disetujui oleh Koordinator Sosial"
	case enum.RoleChairman:
		if existing.Status != StatusSocialManagerAccepted {
			return pkg.NewResponse(http.StatusBadRequest, "Hanya calon yang telah disetujui oleh Koordinator Sosial yang dapat disetujui oleh Ketua Yayasan", nil, nil)
		}
		nextStatus = StatusAccepted
		message = "Calon berhasil disetujui oleh Ketua Yayasan"
	default:
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melakukan tindakan ini", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":           nextStatus,
		"rejection_reason": "",
		"updated_at":       time.Now(),
	}

	if err := s.repo.UpdateFosterChildrenCandidate(ctx, id, updateData); err != nil {
		logrus.WithError(err).Error("failed to update candidate status")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status calon", nil, nil)
	}

	// If fully accepted by chairman, create foster children record
	if nextStatus == StatusAccepted {
		if err := s.fosterChildrenRepo.CreateFosterChildrenFromCandidate(
			ctx,
			existing.ID.String(),
			existing.Name,
			existing.ProfilePicture,
			existing.FamilyCard,
			existing.SKTM,
			existing.BirthPlace,
			existing.SchoolName,
			existing.Address,
			string(existing.Gender),
			string(existing.Category),
			existing.EducationLevel,
			existing.BirthDate,
		); err != nil {
			logrus.WithError(err).Error("failed to create foster children from accepted candidate")
		}
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foster_children_candidate", id, existing.ToFosterChildrenCandidateResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, message, nil, nil)
}

func (s *service) RejectFosterChildrenCandidate(ctx context.Context, id string, req RejectFosterChildrenCandidateRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if req.RejectionReason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"rejectionReason": "Alasan penolakan wajib diisi"}, nil)
	}

	existing, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Calon tidak ditemukan", nil, nil)
	}

	if existing.Status != StatusPending && existing.Status != StatusSocialManagerAccepted {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya calon dengan status pending atau disetujui Koordinator Sosial yang dapat ditolak", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":           StatusRejected,
		"rejection_reason": req.RejectionReason,
		"updated_at":       time.Now(),
	}

	if err := s.repo.UpdateFosterChildrenCandidate(ctx, id, updateData); err != nil {
		logrus.WithError(err).Error("failed to update candidate status")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status calon", nil, nil)
	}

	s.logService.CreateLog(ctx, nil, "UPDATE", "foster_children_candidate", id, existing.ToFosterChildrenCandidateResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Calon berhasil ditolak", nil, nil)
}

func (s *service) CancelFosterChildrenCandidate(ctx context.Context, accountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	existing, err := s.repo.FindOneFosterChildrenCandidate(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Calon tidak ditemukan", nil, nil)
	}

	if existing.SubmittedBy.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk membatalkan calon ini", nil, nil)
	}

	if existing.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya calon dengan status pending yang dapat dibatalkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusCancelled,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateFosterChildrenCandidate(ctx, id, updateData); err != nil {
		logrus.WithError(err).Error("failed to cancel candidate")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membatalkan calon", nil, nil)
	}

	s.logService.CreateLog(ctx, &accountID, "UPDATE", "foster_children_candidate", id, existing.ToFosterChildrenCandidateResponse(), updateData)
	return pkg.NewResponse(http.StatusOK, "Calon berhasil dibatalkan", nil, nil)
}
