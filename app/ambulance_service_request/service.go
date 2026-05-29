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
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	ListMyAmbulanceServiceRequests(ctx context.Context, accountID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	ListAssignedAmbulanceServiceRequests(ctx context.Context, driverAccountID string, ambulanceID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response
	GetAmbulanceServiceRequestByID(ctx context.Context, id string) pkg.Response
	GetMyAmbulanceServiceRequestByID(ctx context.Context, accountID string, id string) pkg.Response
	GetAssignedAmbulanceServiceRequestByID(ctx context.Context, ambulanceID string, id string) pkg.Response
	CreateAmbulanceServiceRequest(ctx context.Context, payload CreateAmbulanceServiceRequest) pkg.Response
	UpdateAmbulanceServiceRequest(ctx context.Context, id string, payload UpdateAmbulanceServiceRequest) pkg.Response
	AcceptAmbulanceServiceRequest(ctx context.Context, id string, role enum.RoleName, payload AcceptAmbulanceServiceRequestPayload) pkg.Response
	RejectAmbulanceServiceRequest(ctx context.Context, id string, req RejectAmbulanceServiceRequest) pkg.Response
	CancelAmbulanceServiceRequest(ctx context.Context, accountID string, id string) pkg.Response
	StartAmbulanceServiceRequest(ctx context.Context, driverAccountID string, ambulanceID string, id string) pkg.Response
	CompleteAmbulanceServiceRequest(ctx context.Context, driverAccountID string, ambulanceID string, id string) pkg.Response
}

type service struct {
	repo                 Repository
	ambulanceRepo        ambulance.Repository
	ambulanceHistoryRepo ambulance_history.Repository
	timeout              time.Duration
}

func NewService(repo Repository, ambulanceRepo ambulance.Repository, ambulanceHistoryRepo ambulance_history.Repository, timeout time.Duration) Service {
	return &service{repo: repo, ambulanceRepo: ambulanceRepo, ambulanceHistoryRepo: ambulanceHistoryRepo, timeout: timeout}
}

func (s *service) ListAmbulanceServiceRequest(ctx context.Context, queryParams AmbulanceServiceRequestQueryParams) pkg.Response {
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
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
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

func (s *service) ListAssignedAmbulanceServiceRequests(ctx context.Context, driverAccountID string, ambulanceID string, queryParams AmbulanceServiceRequestQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": ambulanceID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan ambulans", nil, nil)
	}

	if ambulanceRecord.DriverID.String() != driverAccountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses ke ambulans ini", nil, nil)
	}

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}
	options := map[string]interface{}{
		"limit":        queryParams.Limit,
		"ambulance_id": ambulanceID,
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
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

func (s *service) GetAssignedAmbulanceServiceRequestByID(ctx context.Context, ambulanceID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": ambulanceID})
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
	if payload.ApplicantName == "" {
		errValidation["applicantName"] = "Nama pemohon wajib diisi"
	}
	if payload.ApplicantPhone == "" {
		errValidation["applicantPhone"] = "Nomor telepon pemohon wajib diisi"
	}
	if payload.ApplicantAddress == "" {
		errValidation["applicantAddress"] = "Alamat pemohon wajib diisi"
	}
	if payload.RequestDate == "" {
		errValidation["requestDate"] = "Tanggal permintaan wajib diisi"
	}
	if payload.RequestReason == "" {
		errValidation["requestReason"] = "Alasan permintaan wajib diisi"
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

	request := AmbulanceServiceRequest{
		ID:               uuid.New(),
		AccountID:        uuid.MustParse(payload.AccountID),
		ApplicantName:    payload.ApplicantName,
		ApplicantPhone:   payload.ApplicantPhone,
		ApplicantAddress: payload.ApplicantAddress,
		RequestDate:      time.Now(),
		RequestReason:    payload.RequestReason,
		Status:           StatusPending,
		ServiceCategory:  ambulance_history.ServiceCategory(payload.ServiceCategory),
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat permintaan ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil dibuat", nil, nil)
}

func (s *service) UpdateAmbulanceServiceRequest(ctx context.Context, id string, payload UpdateAmbulanceServiceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}

	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	errValidation := make(map[string]string)
	updateData := make(map[string]interface{})
	if payload.Status != "" && payload.Status != string(StatusPending) && payload.Status != string(StatusApproved) && payload.Status != string(StatusRejected) {
		errValidation["status"] = "Nilai status tidak valid"
	} else if payload.Status != "" {
		updateData["status"] = payload.Status
	}
	if payload.Status == string(StatusRejected) && payload.RejectionReason == "" {
		errValidation["rejectionReason"] = "Alasan penolakan wajib diisi bila status ditolak"
	} else if payload.Status == string(StatusRejected) {
		updateData["rejection_reason"] = payload.RejectionReason
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}
	if len(updateData) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Tidak ada data untuk diperbarui", nil, nil)
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui permintaan ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Permintaan ambulans berhasil diperbarui", nil, nil)
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

	if ambulanceServiceRequest.AccountID.String() != accountID {
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
		"status":           StatusApproved,
		"ambulance_id":     ambulanceRecord.ID,
		"rejection_reason": "",
		"updated_at":       time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
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

	if existing.AccountID.String() != accountID {
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

func (s *service) StartAmbulanceServiceRequest(ctx context.Context, driverAccountID string, ambulanceID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}
	if err := uuid.Validate(ambulanceID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"ambulanceId": "Format ID ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": ambulanceID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengecek data ambulans", nil, nil)
	}

	if ambulanceRecord.DriverID.String() != driverAccountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses ke ambulans ini", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.AmbulanceID == nil || existing.AmbulanceID.String() != ambulanceID {
		return pkg.NewResponse(http.StatusForbidden, "Ambulans ini tidak ditugaskan untuk permintaan ini", nil, nil)
	}

	if existing.Status != StatusApproved {
		return pkg.NewResponse(http.StatusBadRequest, "Hanya permintaan dengan status disetujui yang dapat dimulai", nil, nil)
	}

	updateData := map[string]interface{}{
		"status":     StatusInService,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui status permintaan ambulans", nil, nil)
	}

	// Update ambulance status to "in use"
	_ = s.ambulanceRepo.UpdateAmbulance(ctx, ambulanceID, map[string]interface{}{
		"status":     ambulance.AmbulanceStatusInUse,
		"updated_at": time.Now(),
	})

	return pkg.NewResponse(http.StatusOK, "Layanan ambulans berhasil dimulai", nil, nil)
}

func (s *service) CompleteAmbulanceServiceRequest(ctx context.Context, driverAccountID string, ambulanceID string, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(id); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID permintaan ambulans tidak valid"}, nil)
	}
	if err := uuid.Validate(ambulanceID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"ambulanceId": "Format ID ambulans tidak valid"}, nil)
	}

	ambulanceRecord, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": ambulanceID})
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengecek data ambulans", nil, nil)
	}

	if ambulanceRecord.DriverID.String() != driverAccountID {
		return pkg.NewResponse(http.StatusForbidden, "Anda tidak memiliki akses ke ambulans ini", nil, nil)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Permintaan ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memuat permintaan ambulans", nil, nil)
	}

	if existing.AmbulanceID == nil || existing.AmbulanceID.String() != ambulanceID {
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

	// Update ambulance status to "available"
	_ = s.ambulanceRepo.UpdateAmbulance(ctx, ambulanceID, map[string]interface{}{
		"status":     ambulance.AmbulanceStatusAvailable,
		"updated_at": time.Now(),
	})

	// Automatically create ambulance history
	now := time.Now()
	note := fmt.Sprintf("Layanan ambulans selesai untuk permintaan dari %s. Alasan: %s", existing.ApplicantName, existing.RequestReason)
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
