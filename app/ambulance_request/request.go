package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceRequest struct {
	AccountID        string `json:"accountId"`
	ApplicantName    string `json:"applicantName"`
	ApplicantPhone   string `json:"applicantPhone"`
	ApplicantAddress string `json:"applicantAddress"`
	RequestDate      string `json:"requestDate"`
	RequestReason    string `json:"requestReason"`
}

type UpdateAmbulanceRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}

type AmbulanceRequestQueryParams struct {
	AccountID string `form:"accountId"`
	Status    string `form:"status"`
	pkg.PaginationParams
}
