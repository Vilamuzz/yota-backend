package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceRequest struct {
	UserID           int    `json:"user_id"`
	ApplicantName    string `json:"applicant_name"`
	ApplicantPhone   string `json:"applicant_phone"`
	ApplicantAddress string `json:"applicant_address"`
	Date             string `json:"date"`
	Reason           string `json:"reason"`
}

type UpdateAmbulanceRequest struct {
	Status       string `json:"status"`
	RejectReason string `json:"reject_reason"`
}

type AmbulanceRequestQueryParams struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	pkg.CursorPagination
}
