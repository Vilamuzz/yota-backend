package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type AmbulanceRequestResponse struct {
	ID               string `json:"id"`
	AccountID        string `json:"account_id"`
	ApplicantName    string `json:"applicant_name"`
	ApplicantPhone   string `json:"applicant_phone"`
	ApplicantAddress string `json:"applicant_address"`
	RequestDate      string `json:"request_date"`
	RequestReason    string `json:"request_reason"`
	Status           Status `json:"status"`
	RejectionReason  string `json:"rejection_reason"`
	CreatedAt        string `json:"created_at"`
}

type AmbulanceRequestListResponse struct {
	Requests   []AmbulanceRequestResponse `json:"requests"`
	Pagination pkg.CursorPagination       `json:"pagination"`
}

func (a *AmbulanceRequest) toAmbulanceRequestResponse() AmbulanceRequestResponse {
	return AmbulanceRequestResponse{
		ID:               a.ID.String(),
		AccountID:        a.AccountID.String(),
		ApplicantName:    a.ApplicantName,
		ApplicantPhone:   a.ApplicantPhone,
		ApplicantAddress: a.ApplicantAddress,
		RequestDate:      a.RequestDate.Format("2006-01-02 15:04:05"),
		RequestReason:    a.RequestReason,
		Status:           a.Status,
		RejectionReason:  a.RejectionReason,
		CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAmbulanceRequestsToListResponse(requests []AmbulanceRequest, pagination pkg.CursorPagination) AmbulanceRequestListResponse {
	var responses []AmbulanceRequestResponse
	for _, request := range requests {
		responses = append(responses, request.toAmbulanceRequestResponse())
	}
	if requests == nil {
		responses = []AmbulanceRequestResponse{}
	}
	return AmbulanceRequestListResponse{
		Requests:   responses,
		Pagination: pagination,
	}
}
