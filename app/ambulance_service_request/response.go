package ambulance_service_request

import (
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type AmbulanceServiceRequestResponse struct {
	ID                string                            `json:"id"`
	AccountID         string                            `json:"accountId"`
	ApplicantName     string                            `json:"applicantName"`
	ApplicantPhone    string                            `json:"applicantPhone"`
	ApplicantAddress  string                            `json:"applicantAddress"`
	RequestDate       string                            `json:"requestDate"`
	RequestReason     string                            `json:"requestReason"`
	Status            Status                            `json:"status"`
	ServiceCategory   ambulance_history.ServiceCategory `json:"serviceCategory"`
	RejectionReason   string                            `json:"rejectionReason"`
	AssignedAmbulance *ambulance.AmbulanceResponse      `json:"assignedAmbulance"`
	CreatedAt         string                            `json:"createdAt"`
}

type AmbulanceServiceRequestListResponse struct {
	Requests   []AmbulanceServiceRequestResponse `json:"requests"`
	Pagination pkg.CursorPagination              `json:"pagination"`
}

func (a *AmbulanceServiceRequest) toAmbulanceServiceRequestResponse() AmbulanceServiceRequestResponse {
	resp := AmbulanceServiceRequestResponse{
		ID:               a.ID.String(),
		AccountID:        a.AccountID.String(),
		ApplicantName:    a.ApplicantName,
		ApplicantPhone:   a.ApplicantPhone,
		ApplicantAddress: a.ApplicantAddress,
		RequestDate:      a.RequestDate.Format("2006-01-02 15:04:05"),
		RequestReason:    a.RequestReason,
		Status:           a.Status,
		ServiceCategory:  a.ServiceCategory,
		RejectionReason:  a.RejectionReason,
		CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if a.Ambulance != nil {
		ambulanceResp := a.Ambulance.ToAmbulanceResponse()
		resp.AssignedAmbulance = &ambulanceResp
	}

	return resp
}

func toAmbulanceServiceRequestsToListResponse(requests []AmbulanceServiceRequest, pagination pkg.CursorPagination) AmbulanceServiceRequestListResponse {
	var responses []AmbulanceServiceRequestResponse
	for _, request := range requests {
		responses = append(responses, request.toAmbulanceServiceRequestResponse())
	}
	if requests == nil {
		responses = []AmbulanceServiceRequestResponse{}
	}
	return AmbulanceServiceRequestListResponse{
		Requests:   responses,
		Pagination: pagination,
	}
}

type AmbulanceServiceRequestAdminListResponse struct {
	Requests   []AmbulanceServiceRequestResponse `json:"requests"`
	Pagination pkg.OffsetPagination              `json:"pagination"`
}

func toAmbulanceServiceRequestsToAdminListResponse(requests []AmbulanceServiceRequest, pagination pkg.OffsetPagination) AmbulanceServiceRequestAdminListResponse {
	var responses []AmbulanceServiceRequestResponse
	for _, request := range requests {
		responses = append(responses, request.toAmbulanceServiceRequestResponse())
	}
	if requests == nil {
		responses = []AmbulanceServiceRequestResponse{}
	}
	return AmbulanceServiceRequestAdminListResponse{
		Requests:   responses,
		Pagination: pagination,
	}
}

