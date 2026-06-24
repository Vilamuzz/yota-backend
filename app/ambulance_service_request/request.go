package ambulance_service_request

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateAmbulanceServiceRequest struct {
	AccountID       string                `form:"-"`
	SubmitterName   string                `form:"submitterName"`
	SubmitterPhone  string                `form:"submitterPhone"`
	SubmitterIDCard *multipart.FileHeader `form:"submitterIdCard"`
	PatientName     string                `form:"patientName"`
	PatientAddress  string                `form:"patientAddress"`
	PatientAge      int                   `form:"patientAge"`
	IsInfectious    bool                  `form:"isInfectious"`
	Disease         string                `form:"disease"`
	IsAbleToSit     bool                  `form:"isAbleToSit"`
	PickupDate      string                `form:"pickupDate"`
	PickupTime      string                `form:"pickupTime"`
	Destination     string                `form:"destination"`
	Note            string                `form:"note"`
	ServiceCategory string                `form:"serviceCategory"`
}

type AcceptAmbulanceServiceRequestPayload struct {
	AmbulanceID string `json:"ambulanceId"`
}

type RejectAmbulanceServiceRequest struct {
	RejectionReason string `json:"rejectionReason"`
}

type CancelAmbulanceServiceRequestPayload struct {
	CancelationReason string `json:"cancelationReason"`
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
