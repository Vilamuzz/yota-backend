package ambulance_service_request

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceServiceRequest struct {
	AccountID        string `json:"accountId"`
	ApplicantName    string `json:"applicantName"`
	ApplicantPhone   string `json:"applicantPhone"`
	ApplicantAddress string `json:"applicantAddress"`
	RequestDate      string `json:"requestDate"`
	RequestReason    string `json:"requestReason"`
	ServiceCategory  string `json:"serviceCategory"`
}

type AcceptAmbulanceServiceRequestPayload struct {
	AmbulanceID string `json:"ambulanceId"`
}

type RejectAmbulanceServiceRequest struct {
	RejectionReason string `json:"rejectionReason"`
}

type UpdateAmbulanceServiceRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}

type AmbulanceServiceRequestQueryParams struct {
	Search          string `form:"search"`
	SortBy          string `form:"sortBy"`
	AccountID       string `form:"accountId"`
	Status          string `form:"status"`
	ServiceCategory string `form:"serviceCategory"`
	pkg.PaginationParams
}

type AmbulanceServiceRequestAdminQueryParams struct {
	Status          string `form:"status"`
	AccountID       string `form:"accountId"`
	SortBy          string `form:"sortBy"`
	Search          string `form:"search"`
	ServiceCategory string `form:"serviceCategory"`
	Page            int    `form:"page"`
	Limit           int    `form:"limit"`
}

