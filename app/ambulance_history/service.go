package ambulance_history

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response
	AdminListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response
	AmbulanceHistorySummary(ctx context.Context, ambulanceID string, params AmbulanceSummaryQueryParams) pkg.Response
	AllHistorySummary(ctx context.Context, params AmbulanceSummaryQueryParams) pkg.Response
	HistoryMonthlyTrend(ctx context.Context, params MonthlyTrendQueryParams) pkg.Response
	DriverHistorySummary(ctx context.Context, driverID string, params AmbulanceSummaryQueryParams) pkg.Response
	DriverHistoryMonthlyTrend(ctx context.Context, driverID string, params MonthlyTrendQueryParams) pkg.Response
	DriverListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response
	CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response
	UpdateAmbulanceHistory(ctx context.Context, id string, payload UpdateAmbulanceHistoryRequest) pkg.Response
	DeleteAmbulanceHistory(ctx context.Context, id string) pkg.Response
}

type service struct {
	repo          Repository
	ambulanceRepo ambulance.Repository
	timeout       time.Duration
}

func NewService(repo Repository, ambulanceRepo ambulance.Repository, timeout time.Duration) Service {
	return &service{
		repo:          repo,
		ambulanceRepo: ambulanceRepo,
		timeout:       timeout,
	}
}

func (s *service) ListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.AmbulanceID != "" {
		options["ambulance_id"] = queryParams.AmbulanceID
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	histories, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(500, "Gagal mendapatkan riwayat ambulans", nil, nil)
	}

	hasNext := len(histories) > queryParams.Limit
	if hasNext {
		histories = histories[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(histories) > 0 {
		lastHistory := histories[len(histories)-1]
		nextCursor = pkg.EncodeCursor(lastHistory.CreatedAt, lastHistory.ID.String())
	}
	if hasPrev && len(histories) > 0 {
		firstHistory := histories[0]
		prevCursor = pkg.EncodeCursor(firstHistory.CreatedAt, firstHistory.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, toAmbulanceHistoriesToListResponse(histories, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) AdminListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.AmbulanceID != "" {
		options["ambulance_id"] = queryParams.AmbulanceID
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	histories, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(500, "Gagal mendapatkan riwayat ambulans", nil, nil)
	}

	hasNext := len(histories) > queryParams.Limit
	if hasNext {
		histories = histories[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(histories) > 0 {
		lastHistory := histories[len(histories)-1]
		nextCursor = pkg.EncodeCursor(lastHistory.CreatedAt, lastHistory.ID.String())
	}
	if hasPrev && len(histories) > 0 {
		firstHistory := histories[0]
		prevCursor = pkg.EncodeCursor(firstHistory.CreatedAt, firstHistory.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, toAmbulanceHistoriesToAdminListResponse(histories, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}

func (s *service) AmbulanceHistorySummary(ctx context.Context, ambulanceID string, params AmbulanceSummaryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if ambulanceID == "" {
		return pkg.NewResponse(http.StatusBadRequest, "ID Ambulans wajib diisi", nil, nil)
	}

	var startDate, endDate *time.Time
	now := time.Now()

	if params.StartDate != "" {
		parsedStart, err := time.ParseInLocation("2006-01-02", params.StartDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format startDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.StartDate), nil, nil)
		}
		startDate = &parsedStart
	}

	if params.EndDate != "" {
		parsedEnd, err := time.ParseInLocation("2006-01-02", params.EndDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format endDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.EndDate), nil, nil)
		}
		parsedEnd = parsedEnd.Add(24*time.Hour - time.Nanosecond) // inclusive end
		endDate = &parsedEnd
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return pkg.NewResponse(http.StatusBadRequest, "startDate harus sebelum endDate", nil, nil)
	}

	counts, err := s.repo.GetSummary(ctx, ambulanceID, startDate, endDate)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan ringkasan riwayat ambulans", nil, nil)
	}

	var total int64
	for _, c := range counts {
		total += c.Count
	}

	summary := SummaryResponse{
		Total:      total,
		Categories: counts,
	}
	if startDate != nil {
		summary.StartDate = startDate.Format("2006-01-02")
	}
	if endDate != nil {
		summary.EndDate = endDate.Format("2006-01-02")
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, summary)
}

func (s *service) AllHistorySummary(ctx context.Context, params AmbulanceSummaryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var startDate, endDate *time.Time
	now := time.Now()

	if params.StartDate != "" {
		parsedStart, err := time.ParseInLocation("2006-01-02", params.StartDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format startDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.StartDate), nil, nil)
		}
		startDate = &parsedStart
	}

	if params.EndDate != "" {
		parsedEnd, err := time.ParseInLocation("2006-01-02", params.EndDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format endDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.EndDate), nil, nil)
		}
		parsedEnd = parsedEnd.Add(24*time.Hour - time.Nanosecond) // inclusive end
		endDate = &parsedEnd
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return pkg.NewResponse(http.StatusBadRequest, "startDate harus sebelum endDate", nil, nil)
	}

	counts, err := s.repo.GetAllSummary(ctx, startDate, endDate)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan semua ringkasan riwayat ambulans", nil, nil)
	}

	var total int64
	for _, c := range counts {
		total += c.Count
	}

	summary := SummaryResponse{
		Total:      total,
		Categories: counts,
	}
	if startDate != nil {
		summary.StartDate = startDate.Format("2006-01-02")
	}
	if endDate != nil {
		summary.EndDate = endDate.Format("2006-01-02")
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, summary)
}

func (s *service) HistoryMonthlyTrend(ctx context.Context, params MonthlyTrendQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	yearVal := time.Now().Year()
	if params.Year != "" {
		var parseYear int
		if _, err := fmt.Sscanf(params.Year, "%d", &parseYear); err == nil && parseYear > 0 {
			yearVal = parseYear
		}
	}

	trend, err := s.repo.GetMonthlyTrend(ctx, yearVal)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan tren bulanan riwayat ambulans", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, trend)
}

func (s *service) CreateAmbulanceHistory(ctx context.Context, payload CreateAmbulanceHistoryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.AmbulanceID == "" {
		errValidation["ambulance_id"] = "ID Ambulans wajib diisi"
	} else if payload.AmbulanceID != "" {
		_, err := s.ambulanceRepo.FindOneAmbulance(ctx, map[string]interface{}{"id": payload.AmbulanceID})
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				errValidation["ambulance_id"] = "Ambulans tidak ditemukan"
			} else {
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan ambulans", nil, nil)
			}
		}
	}
	if payload.ServiceCategory == "" {
		errValidation["service_category"] = "Kategori layanan wajib diisi"
	} else if payload.ServiceCategory != SocialService && payload.ServiceCategory != EmergencyService && payload.ServiceCategory != OtherService && payload.ServiceCategory != MortuaryService && payload.ServiceCategory != PatientService {
		errValidation["service_category"] = "Kategori layanan tidak valid"
	}

	if payload.DriverID == "" {
		errValidation["driver_id"] = "ID Supir wajib diisi"
	} else if _, err := uuid.Parse(payload.DriverID); err != nil {
		errValidation["driver_id"] = "Format ID Supir tidak valid"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	now := time.Now()
	ambulanceHistory := AmbulanceHistory{
		ID:              uuid.New(),
		AmbulanceID:     uuid.MustParse(payload.AmbulanceID),
		DriverID:        uuid.MustParse(payload.DriverID),
		ServiceCategory: payload.ServiceCategory,
		Note:            payload.Note,
		CreatedAt:       now,
	}
	if err := s.repo.Create(ctx, ambulanceHistory); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat riwayat ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusCreated, "Riwayat ambulans berhasil dibuat", nil, nil)
}

func (s *service) UpdateAmbulanceHistory(ctx context.Context, id string, payload UpdateAmbulanceHistoryRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.ServiceCategory == "" {
		errValidation["service_category"] = "Kategori layanan wajib diisi"
	} else if payload.ServiceCategory != SocialService && payload.ServiceCategory != EmergencyService && payload.ServiceCategory != OtherService && payload.ServiceCategory != MortuaryService && payload.ServiceCategory != PatientService {
		errValidation["service_category"] = "Kategori layanan tidak valid"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	ambulanceHistory, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Riwayat ambulans tidak ditemukan", nil, nil)
	}

	ambulanceHistory.ServiceCategory = payload.ServiceCategory
	ambulanceHistory.Note = payload.Note

	if err := s.repo.Update(ctx, ambulanceHistory); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui riwayat ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Riwayat ambulans berhasil diperbarui", nil, nil)
}

func (s *service) DeleteAmbulanceHistory(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return pkg.NewResponse(http.StatusNotFound, "Riwayat ambulans tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan riwayat ambulans", nil, nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus riwayat ambulans", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Riwayat ambulans berhasil dihapus", nil, nil)
}

func (s *service) DriverHistorySummary(ctx context.Context, driverID string, params AmbulanceSummaryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if driverID == "" {
		return pkg.NewResponse(http.StatusBadRequest, "ID Supir wajib diisi", nil, nil)
	}

	var startDate, endDate *time.Time
	now := time.Now()

	if params.StartDate != "" {
		parsedStart, err := time.ParseInLocation("2006-01-02", params.StartDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format startDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.StartDate), nil, nil)
		}
		startDate = &parsedStart
	}

	if params.EndDate != "" {
		parsedEnd, err := time.ParseInLocation("2006-01-02", params.EndDate, now.Location())
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest,
				fmt.Sprintf("format endDate tidak valid: %s (diharapkan YYYY-MM-DD)", params.EndDate), nil, nil)
		}
		parsedEnd = parsedEnd.Add(24*time.Hour - time.Nanosecond) // inclusive end
		endDate = &parsedEnd
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return pkg.NewResponse(http.StatusBadRequest, "startDate harus sebelum endDate", nil, nil)
	}

	counts, err := s.repo.GetDriverSummary(ctx, driverID, startDate, endDate)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan ringkasan riwayat ambulans supir", nil, nil)
	}

	var total int64
	for _, c := range counts {
		total += c.Count
	}

	summary := SummaryResponse{
		Total:      total,
		Categories: counts,
	}
	if startDate != nil {
		summary.StartDate = startDate.Format("2006-01-02")
	}
	if endDate != nil {
		summary.EndDate = endDate.Format("2006-01-02")
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, summary)
}

func (s *service) DriverHistoryMonthlyTrend(ctx context.Context, driverID string, params MonthlyTrendQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if driverID == "" {
		return pkg.NewResponse(http.StatusBadRequest, "ID Supir wajib diisi", nil, nil)
	}

	yearVal := time.Now().Year()
	if params.Year != "" {
		var parseYear int
		if _, err := fmt.Sscanf(params.Year, "%d", &parseYear); err == nil && parseYear > 0 {
			yearVal = parseYear
		}
	}

	trend, err := s.repo.GetDriverMonthlyTrend(ctx, driverID, yearVal)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan tren bulanan riwayat ambulans supir", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, trend)
}

func (s *service) DriverListAmbulanceHistory(ctx context.Context, queryParams AmbulanceHistoryQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}
	if queryParams.AmbulanceID != "" {
		options["ambulance_id"] = queryParams.AmbulanceID
	}
	if queryParams.DriverID != "" {
		options["driver_id"] = queryParams.DriverID
	}
	if queryParams.ServiceCategory != "" {
		options["service_category"] = queryParams.ServiceCategory
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if queryParams.PrevCursor != "" {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	histories, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mendapatkan riwayat ambulans supir", nil, nil)
	}

	hasNext := len(histories) > queryParams.Limit
	if hasNext {
		histories = histories[:queryParams.Limit]
	}

	var nextCursor, prevCursor string
	hasPrev := queryParams.PrevCursor != ""
	if hasNext && len(histories) > 0 {
		lastHistory := histories[len(histories)-1]
		nextCursor = pkg.EncodeCursor(lastHistory.CreatedAt, lastHistory.ID.String())
	}
	if hasPrev && len(histories) > 0 {
		firstHistory := histories[0]
		prevCursor = pkg.EncodeCursor(firstHistory.CreatedAt, firstHistory.ID.String())
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, toAmbulanceHistoriesToAdminListResponse(histories, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      queryParams.Limit,
	}))
}
