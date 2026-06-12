package ambulance_service_request

import (
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type AmbulanceServiceRequestResponse struct {
	ID              string                            `json:"id"`
	SubmittedBy     string                            `json:"submittedBy"`
	SubmitterName   string                            `json:"submitterName"`
	SubmitterPhone  string                            `json:"submitterPhone"`
	SubmitterIDCard string                            `json:"submitterIdCard"`
	PatientName     string                            `json:"patientName"`
	PatientAddress  string                            `json:"patientAddress"`
	PatientAge      int                               `json:"patientAge"`
	IsInfectious    bool                              `json:"isInfectious"`
	Disease         string                            `json:"disease"`
	IsAbleToSit     bool                              `json:"isAbleToSit"`
	PickupDate      string                            `json:"pickupDate"`
	PickupTime      string                            `json:"pickupTime"`
	Destination     string                            `json:"destination"`
	Note            string                            `json:"note"`
	Status          Status                            `json:"status"`
	ServiceCategory ambulance_history.ServiceCategory `json:"serviceCategory"`
	RejectionReason string                            `json:"rejectionReason"`
	AssignedAmbulance *ambulance.AmbulanceResponse    `json:"assignedAmbulance"`
	CreatedAt       string                            `json:"createdAt"`
}

type AmbulanceServiceRequestListResponse struct {
	Requests   []AmbulanceServiceRequestResponse `json:"requests"`
	Pagination pkg.CursorPagination              `json:"pagination"`
}

func (a *AmbulanceServiceRequest) toAmbulanceServiceRequestResponse() AmbulanceServiceRequestResponse {
	resp := AmbulanceServiceRequestResponse{
		ID:              a.ID.String(),
		SubmittedBy:     a.SubmittedBy.String(),
		SubmitterName:   a.SubmitterName,
		SubmitterPhone:  a.SubmitterPhone,
		SubmitterIDCard: a.SubmitterIDCard,
		PatientName:     a.PatientName,
		PatientAddress:  a.PatientAddress,
		PatientAge:      a.PatientAge,
		IsInfectious:    a.IsInfectious,
		Disease:         a.Disease,
		IsAbleToSit:     a.IsAbleToSit,
		PickupDate:      a.PickupDate.Format("2006-01-02"),
		PickupTime:      a.PickupTime.Format("15:04:05"),
		Destination:     a.Destination,
		Note:            a.Note,
		Status:          a.Status,
		ServiceCategory: a.ServiceCategory,
		RejectionReason: a.RejectionReason,
		CreatedAt:       a.CreatedAt.Format("2006-01-02 15:04:05"),
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
