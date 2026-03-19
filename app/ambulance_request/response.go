package ambulance_request

import "github.com/Vilamuzz/yota-backend/pkg"

type AmbulanceRequestResponse struct {
	ID               string `json:"id"`
	UserID           int    `json:"user_id"`
	ApplicantName    string `json:"applicant_name"`
	ApplicantPhone   string `json:"applicant_phone"`
	ApplicantAddress string `json:"applicant_address"`
	Date             string `json:"date"`
	Reason           string `json:"reason"`
	Status           Status `json:"status"`
	RejectReason     string `json:"reject_reason"`
	CreatedAt        string `json:"created_at"`
}

type AmbulanceRequestListResponse struct {
	Requests   []AmbulanceRequestResponse `json:"requests"`
	Pagination pkg.CursorPagination       `json:"pagination"`
}

func (a *AmbulanceRequest) toAmbulanceRequestResponse() AmbulanceRequestResponse {
	return AmbulanceRequestResponse{
		ID:               a.ID,
		UserID:           a.UserID,
		ApplicantName:    a.ApplicantName,
		ApplicantPhone:   a.ApplicantPhone,
		ApplicantAddress: a.ApplicantAddress,
		Date:             a.Date.Format("2006-01-02 15:04:05"),
		Reason:           a.Reason,
		Status:           a.Status,
		RejectReason:     a.RejectReason,
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
