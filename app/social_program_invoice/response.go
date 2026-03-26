package social_program_invoice

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramInvoiceResponse struct {
	ID             string  `json:"id"`
	SubscriptionID string  `json:"subscription_id"`
	Year           int     `json:"year"`
	Month          int     `json:"month"`
	MinimumAmount  float64 `json:"minimum_amount"`
	Status         string  `json:"status"`
}

type SocialProgramInvoiceListResponse struct {
	Invoices   []SocialProgramInvoiceResponse `json:"invoices"`
	Pagination pkg.CursorPagination           `json:"pagination"`
}

func (r *SocialProgramInvoice) toSocialProgramInvoiceResponse() SocialProgramInvoiceResponse {
	return SocialProgramInvoiceResponse{
		ID:             r.ID,
		SubscriptionID: r.SubscriptionID,
		Year:           r.Year,
		Month:          r.Month,
		MinimumAmount:  r.MinimumAmount,
		Status:         string(r.Status),
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
