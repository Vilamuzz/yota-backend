package social_program_invoice

import "github.com/Vilamuzz/yota-backend/pkg"

type SocialProgramInvoiceResponse struct {
	ID             string  `json:"id"`
	SubscriptionID string  `json:"subscriptionId"`
	BillingPeriod  string  `json:"billingPeriod"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
	DueDate        string  `json:"dueDate"`
}

type SocialProgramInvoiceListResponse struct {
	Invoices   []SocialProgramInvoiceResponse `json:"invoices"`
	Pagination pkg.CursorPagination           `json:"pagination"`
}

func (r *SocialProgramInvoice) toSocialProgramInvoiceResponse() SocialProgramInvoiceResponse {
	return SocialProgramInvoiceResponse{
		ID:             r.ID.String(),
		SubscriptionID: r.SubscriptionID.String(),
		BillingPeriod:  r.BillingPeriod.Format("2006-01-02 15:04:05"),
		Amount:         r.Amount,
		Status:         string(r.Status),
		DueDate:        r.DueDate.Format("2006-01-02 15:04:05"),
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
