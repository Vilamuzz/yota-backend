package ambulance_service_request

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceServiceRequest struct {
	AccountID        string `json:"accountId"`
	ApplicantName    string `json:"applicantName"`
	ApplicantPhone   string `json:"applicantPhone"`
	ApplicantAddress string `json:"applicantAddress"`
	RequestDate      string `json:"requestDate"`
	RequestReason    string `json:"requestReason"`
}

type UpdateAmbulanceServiceRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}

type AmbulanceServiceRequestQueryParams struct {
	AccountID string `form:"accountId"`
	Status    string `form:"status"`
	pkg.PaginationParams
}
