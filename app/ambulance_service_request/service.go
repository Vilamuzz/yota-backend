package ambulance_service_request

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestAdminQueryParams) pkg.Response
	ListMyAmbulanceServiceRequests(ctx context.Context, accountID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	ListAssignedAmbulanceServiceRequests(ctx context.Context, driverAccountID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	GetAmbulanceServiceRequestByID(ctx context.Context, id string) pkg.Response
	GetMyAmbulanceServiceRequestByID(ctx context.Context, accountID string, id string) pkg.Response
	GetAssignedAmbulanceServiceRequestByID(ctx context.Context, driverAccountID string, id string) pkg.Response
	CreateAmbulanceServiceRequest(ctx context.Context, payload CreateAmbulanceServiceRequest) pkg.Response
	AcceptAmbulanceServiceRequest(ctx context.Context, id string, role enum.RoleName, payload AcceptAmbulanceServiceRequestPayload) pkg.Response
	RejectAmbulanceServiceRequest(ctx context.Context, id string, req RejectAmbulanceServiceRequest) pkg.Response
	CancelAmbulanceServiceRequest(ctx context.Context, accountID string, id string) pkg.Response
	DriverCancelAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string, payload CancelAmbulanceServiceRequestPayload) pkg.Response
	StartAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string) pkg.Response
	CompleteAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string) pkg.Response
}

type service struct {
	repo                 Repository
	ambulanceRepo        ambulance.Repository
	ambulanceHistoryRepo ambulance_history.Repository
	timeout              time.Duration
	emailService         *pkg.EmailService
	s3Client             s3_pkg.Client
}

func NewService(repo Repository, ambulanceRepo ambulance.Repository, ambulanceHistoryRepo ambulance_history.Repository, timeout time.Duration, s3Client s3_pkg.Client) Service {
	return &service{
		repo:                 repo,
		ambulanceRepo:        ambulanceRepo,
		ambulanceHistoryRepo: ambulanceHistoryRepo,
		timeout:              timeout,
		emailService:         pkg.NewEmailService(),
		s3Client:             s3Client,
	}
}

func (s *service) ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestAdminQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit <= 0 {
		queryParams.Limit = 10
	}
	if queryParams.Limit > 100 {
		queryParams.Limit = 100
	}
	if queryParams.Page <= 0 {
		queryParams.Page = 1
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
		"page":  queryParams.Page,
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.AccountID != "" {
		options["account_id"] = queryParams.AccountID
	}
	if queryParams.SortBy != "" {
		options["sort_by"] = queryParams.SortBy
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}

	total, err := s.repo.Count(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil jumlah total permintaan ambulans", nil, nil)
	}

	ambulanceServiceRequests, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat daftar permintaan ambulans", nil, nil)
	}

	totalPages := int((total + int64(queryParams.Limit) - 1) / int64(queryParams.Limit))
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := pkg.OffsetPagination{
		Page:       queryParams.Page,
		Limit:      queryParams.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toAmbulanceServiceRequestsToAdminListResponse(ambulanceServiceRequests, pagination))
}

func (s *service) ListMyAmbulanceServiceRequests(ctx context.Context, accountID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}
	options := map[string]interface{}{
		"limit":      queryParams.Limit,
		"account_id": accountID,
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
		options["search_only_name"] = true
	}
	if queryParams.SortBy != "" {
		options["sort_by"] = queryParams.SortBy
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}

	ambulanceServiceRequests, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat daftar permintaan ambulans", nil, nil)
	}
	hasNext := len(ambulanceServiceRequests) > queryParams.Limit
	if hasNext {
		ambulanceServiceRequests = ambulanceServiceRequests[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulanceServiceRequests) > 0 {
		lastRequest := ambulanceServiceRequests[len(ambulanceServiceRequests)-1]
		nextCursor = pkg.EncodeCursor(lastRequest.CreatedAt, lastRequest.ID.String())
	}
	if hasPrev && len(ambulanceServiceRequests) > 0 {
		firstRequest := ambulanceServiceRequests[0]
		prevCursor = pkg.EncodeCursor(firstRequest.CreatedAt, firstRequest.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toAmbulanceServiceRequestsToListResponse(ambulanceServiceRequests, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) ListAssignedAmbulanceServiceRequests(ctx context.Context, driverAccountID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"driver_id": driverAccountID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Tidak ada ambulans yang ditugaskan ke supir ini", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan ambulans supir ini", nil, nil)
	}

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}
	options := map[string]interface{}{
		"limit":        queryParams.Limit,
		"ambulance_id": ambulanceRecord.ID.String(),
	}
	if queryParams.Status != "" {
		options["status"] = queryParams.Status
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
		options["search_only_name"] = true
	}
	if queryParams.SortBy != "" {
		options["sort_by"] = queryParams.SortBy
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}

	ambulanceServiceRequests, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat daftar permintaan ambulans yang ditugaskan", nil, nil)
	}
	hasNext := len(ambulanceServiceRequests) > queryParams.Limit
	if hasNext {
		ambulanceServiceRequests = ambulanceServiceRequests[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(ambulanceServiceRequests) > 0 {
		lastRequest := ambulanceServiceRequests[len(ambulanceServiceRequests)-1]
		nextCursor = pkg.EncodeCursor(lastRequest.CreatedAt, lastRequest.ID.String())
	}
	if hasPrev && len(ambulanceServiceRequests) > 0 {
		firstRequest := ambulanceServiceRequests[0]
		prevCursor = pkg.EncodeCursor(firstRequest.CreatedAt, firstRequest.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, toAmbulanceServiceRequestsToListResponse(ambulanceServiceRequests, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) GetAssignedAmbulanceServiceRequestByID(ctx context.Context, driverAccountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"driver_id": driverAccountID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Tidak ada ambulans yang ditugaskan ke supir ini", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan ambulans supir ini", nil, nil)
	}

	ambulanceServiceRequest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if ambulanceServiceRequest.AmbulanceID == nil || *ambulanceServiceRequest.AmbulanceID != ambulanceRecord.ID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melihat permintaan ini", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ambulanceServiceRequest.toAmbulanceServiceRequestResponse())
}

func (s *service) GetAmbulanceServiceRequestByID(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ambulanceServiceRequest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ambulanceServiceRequest.toAmbulanceServiceRequestResponse())
}

func (s *service) CreateAmbulanceServiceRequest(ctx context.Context, payload CreateAmbulanceServiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.SubmitterName == "" {
		errValidation["submitterName"] = "Nama pengirim wajib diisi"
	}
	if payload.SubmitterPhone == "" {
		errValidation["submitterPhone"] = "Nomor telepon pengirim wajib diisi"
	}
	if payload.SubmitterIDCard == nil {
		errValidation["submitterIdCard"] = "KTP pengirim wajib diisi"
	}
	if payload.PatientName == "" {
		errValidation["patientName"] = "Nama pasien wajib diisi"
	}
	if payload.PatientAddress == "" {
		errValidation["patientAddress"] = "Alamat pasien wajib diisi"
	}
	if payload.PickupDate == "" {
		errValidation["pickupDate"] = "Tanggal penjemputan wajib diisi"
	}
	if payload.PickupTime == "" {
		errValidation["pickupTime"] = "Jam penjemputan wajib diisi"
	}
	if payload.Destination == "" {
		errValidation["destination"] = "Tujuan wajib diisi"
	}
	if payload.ServiceCategory == "" {
		errValidation["serviceCategory"] = "Kategori layanan wajib diisi"
	} else {
		cat := ambulance_history.ServiceCategory(payload.ServiceCategory)
		if cat != ambulance_history.SocialService &&
			cat != ambulance_history.MortuaryService &&
			cat != ambulance_history.PatientService &&
			cat != ambulance_history.EmergencyService &&
			cat != ambulance_history.OtherService {
			errValidation["serviceCategory"] = "Kategori layanan tidak valid"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	pickupDate, err := time.Parse("2006-01-02", payload.PickupDate)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"pickupDate": "Format tanggal tidak valid, diharapkan YYYY-MM-DD"}, nil)
	}
	pickupTime, err := time.Parse("15:04", payload.PickupTime)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"pickupTime": "Format jam tidak valid, diharapkan HH:MM"}, nil)
	}

	submitterIDCardURL, err := s.s3Client.UploadFile(ctx, payload.SubmitterIDCard, "ambulance-service-requests")
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah KTP pengirim", nil, nil)
	}

	request := AmbulanceServiceRequest{
		ID:              uuid.New(),
		SubmittedBy:     uuid.MustParse(payload.AccountID),
		SubmitterName:   payload.SubmitterName,
		SubmitterPhone:  payload.SubmitterPhone,
		SubmitterIDCard: submitterIDCardURL,
		PatientName:     payload.PatientName,
		PatientAddress:  payload.PatientAddress,
		PatientAge:      payload.PatientAge,
		IsInfectious:    payload.IsInfectious,
		Disease:         payload.Disease,
		IsAbleToSit:     payload.IsAbleToSit,
		PickupDate:      pickupDate,
		PickupTime:      pickupTime,
		Destination:     payload.Destination,
		Note:            payload.Note,
		Status:          StatusPending,
		ServiceCategory: ambulance_history.ServiceCategory(payload.ServiceCategory),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat permintaan ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil dibuat", nil, nil)
}

func (s *service) GetMyAmbulanceServiceRequestByID(ctx context.Context, accountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	ambulanceServiceRequest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if ambulanceServiceRequest.SubmittedBy.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melihat permintaan ini", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil", nil, ambulanceServiceRequest.toAmbulanceServiceRequestResponse())
}

func (s *service) AcceptAmbulanceServiceRequest(ctx context.Context, id string, role enum.RoleName, payload AcceptAmbulanceServiceRequestPayload) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}
	if err := uuid.Validate(payload.AmbulanceID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"ambulanceId": "Format ID ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": payload.AmbulanceID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengecek data ambulans", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	switch role {
	case enum.RoleAmbulanceManager:
		if existing.Status != StatusPending {
			return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status pending yang dapat disetujui", nil, nil)
		}
	default:
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melakukan tindakan ini", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":           StatusAccepted,
		"ambulance_id":     ambulanceRecord.ID,
		"rejection_reason": "",
		"updated_at":       time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
	}

	if existing.Account.Email != "" {
		go func(email, username, submitterName string) {
			if err := s.emailService.SendAmbulanceServiceRequestAcceptedEmail(email, username, submitterName); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":      "ambulance_service_request.service",
					"email":          email,
					"submitter_name": submitterName,
				}).WithError(err).Error("failed to send ambulance request accepted email asynchronously")
			}
		}(existing.Account.Email, existing.Account.UserProfile.Username, existing.SubmitterName)
	}

	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil disetujui", nil, nil)
}

func (s *service) RejectAmbulanceServiceRequest(ctx context.Context, id string, req RejectAmbulanceServiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if req.RejectionReason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"rejectionReason": "Alasan penolakan wajib diisi"}, nil)
	}

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status pending yang dapat ditolak", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":           StatusRejected,
		"rejection_reason": req.RejectionReason,
		"updated_at":       time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
	}

	if existing.Account.Email != "" {
		go func(email, username, submitterName, rejectionReason string) {
			if err := s.emailService.SendAmbulanceServiceRequestRejectedEmail(email, username, submitterName, rejectionReason); err != nil {
				logrus.WithFields(logrus.Fields{
					"component":      "ambulance_service_request.service",
					"email":          email,
					"submitter_name": submitterName,
				}).WithError(err).Error("failed to send ambulance request rejected email asynchronously")
			}
		}(existing.Account.Email, existing.Account.UserProfile.Username, existing.SubmitterName, req.RejectionReason)
	}

	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil ditolak", nil, nil)
}

func (s *service) CancelAmbulanceServiceRequest(ctx context.Context, accountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.SubmittedBy.String() != accountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses untuk melakukan tindakan ini", nil, nil)
	}

	if existing.Status != StatusPending {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status pending yang dapat dibatalkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusCancelled,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membatalkan permintaan ambulans", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil dibatalkan", nil, nil)
}

func (s *service) DriverCancelAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string, payload CancelAmbulanceServiceRequestPayload) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	if payload.CancelationReason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"cancelationReason": "Alasan pembatalan wajib diisi"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"driver_id": driverAccountID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Tidak ada ambulans yang ditugaskan ke supir ini", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan ambulans supir ini", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.AmbulanceID == nil || existing.AmbulanceID.String() != ambulanceRecord.ID.String() {
		return pkg.NewResponse(http.StatusForbidden, "Ambulans ini tidak ditugaskan untuk permintaan ini", nil, nil)
	}

	if existing.Status != StatusAccepted && existing.Status != StatusInService {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status disetujui atau dalam pelayanan yang dapat dibatalkan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":             StatusCancelled,
		"cancelation_reason": payload.CancelationReason,
		"updated_at":         time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membatalkan permintaan ambulans", nil, nil)
	}

	_ = s.ambulanceRepo.UpdateAmbulance(ctx, ambulanceRecord.ID.String(), map[string]interface{}{
		"status":     ambulance.AmbulanceStatusAvailable,
		"updated_at": time.Now(),
	})

	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil dibatalkan oleh supir", nil, nil)
}

func (s *service) StartAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"driver_id": driverAccountID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Tidak ada ambulans yang ditugaskan ke supir ini", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan ambulans supir ini", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.AmbulanceID == nil || existing.AmbulanceID.String() != ambulanceRecord.ID.String() {
		return pkg.NewResponse(http.StatusForbidden, "Ambulans ini tidak ditugaskan untuk permintaan ini", nil, nil)
	}

	if existing.Status != StatusAccepted {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status disetujui yang dapat dimulai", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusInService,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
	}

	_ = s.ambulanceRepo.UpdateAmbulance(ctx, ambulanceRecord.ID.String(), map[string]interface{}{
		"status":     ambulance.AmbulanceStatusInUse,
		"updated_at": time.Now(),
	})

	return pkg.NewResponse(http.StatusOK, "Layanan ambulans berhasil dimulai", nil, nil)
}

func (s *service) CompleteAmbulanceServiceRequest(ctx context.Context, driverAccountID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"driver_id": driverAccountID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Tidak ada ambulans yang ditugaskan ke supir ini", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengecek data ambulans supir ini", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.AmbulanceID == nil || existing.AmbulanceID.String() != ambulanceRecord.ID.String() {
		return pkg.NewResponse(http.StatusForbidden, "Ambulans ini tidak ditugaskan untuk permintaan ini", nil, nil)
	}

	if existing.Status != StatusInService {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status dalam pelayanan yang dapat diselesaikan", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusDone,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
	}

	_ = s.ambulanceRepo.UpdateAmbulance(ctx, ambulanceRecord.ID.String(), map[string]interface{}{
		"status":     ambulance.AmbulanceStatusAvailable,
		"updated_at": time.Now(),
	})

	now := time.Now()
	note := fmt.Sprintf("Layanan ambulans selesai untuk permintaan dari %s. Pasien: %s", existing.SubmitterName, existing.PatientName)
	history := ambulance_history.AmbulanceHistory{
		ID:              uuid.New(),
		AmbulanceID:     ambulanceRecord.ID,
		DriverID:        ambulanceRecord.DriverID,
		ServiceCategory: existing.ServiceCategory,
		Note:            note,
		CreatedAt:       now,
	}

	if err := s.ambulanceHistoryRepo.Create(ctx, history); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Layanan ambulans diselesaikan tetapi gagal membuat riwayat ambulans", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Layanan ambulans berhasil diselesaikan dan riwayat berhasil dibuat", nil, nil)
}
