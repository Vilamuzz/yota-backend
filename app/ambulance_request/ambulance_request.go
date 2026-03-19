package ambulance_request

import "time"

type AmbulanceRequest struct {
	ID               string    `json:"id"`
	UserID           int       `json:"user_id"`
	ApplicantName    string    `json:"applicant_name"`
	ApplicantPhone   string    `json:"applicant_phone"`
	ApplicantAddress string    `json:"applicant_address"`
	Date             time.Time `json:"date"`
	Reason           string    `json:"reason"`
	Status           Status    `json:"status"`
	RejectReason     string    `json:"reject_reason"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)
