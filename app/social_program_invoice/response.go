package social_program_invoice

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type SocialProgramInvoiceResponse struct {
	ID                 string        `json:"id"`
	SocialProgramTitle string        `json:"socialProgramTitle"`
	BillingPeriod      time.Time     `json:"billingPeriod"`
	MinimumAmount      float64       `json:"minimumAmount"`
	Status             InvoiceStatus `json:"status"`
	DueDate            time.Time     `json:"dueDate"`
	SnapToken          string        `json:"snapToken"`
	CreatedAt          time.Time     `json:"createdAt"`
}

type SocialProgramInvoiceListResponse struct {
	Invoices   []SocialProgramInvoiceResponse `json:"invoices"`
	Pagination pkg.CursorPagination           `json:"pagination"`
}

func (r *SocialProgramInvoice) toSocialProgramInvoiceResponse() SocialProgramInvoiceResponse {
	return SocialProgramInvoiceResponse{
		ID:            r.ID.String(),
		BillingPeriod: r.BillingPeriod,
		MinimumAmount: r.MinimumAmount,
		Status:        r.Status,
		DueDate:       r.DueDate,
		SnapToken:     r.SnapToken,
		CreatedAt:     r.CreatedAt,
	}
}

func toSocialProgramInvoiceListResponse(invoices []SocialProgramInvoice, pagination pkg.CursorPagination) SocialProgramInvoiceListResponse {
	var responses []SocialProgramInvoiceResponse
	for _, invoice := range invoices {
		responses = append(responses, invoice.toSocialProgramInvoiceResponse())
	}
	if responses == nil {
		responses = []SocialProgramInvoiceResponse{}
	}
	return SocialProgramInvoiceListResponse{
		Invoices:   responses,
		Pagination: pagination,
	}
}
