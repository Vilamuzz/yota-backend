package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceRequest struct {
	AccountID        string `json:"account_id"`
	ApplicantName    string `json:"applicant_name"`
	ApplicantPhone   string `json:"applicant_phone"`
	ApplicantAddress string `json:"applicant_address"`
	RequestDate      string `json:"request_date"`
	RequestReason    string `json:"request_reason"`
}

type UpdateAmbulanceRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejection_reason"`
}

type AmbulanceRequestQueryParams struct {
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
	pkg.CursorPagination
}
