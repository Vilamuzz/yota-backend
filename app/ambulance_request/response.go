package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type AmbulanceRequestResponse struct {
	ID               string `json:"id"`
	AccountID        string `json:"accountId"`
	ApplicantName    string `json:"applicantName"`
	ApplicantPhone   string `json:"applicantPhone"`
	ApplicantAddress string `json:"applicantAddress"`
	RequestDate      string `json:"requestDate"`
	RequestReason    string `json:"requestReason"`
	Status           Status `json:"status"`
	RejectionReason  string `json:"rejectionReason"`
	CreatedAt        string `json:"createdAt"`
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
