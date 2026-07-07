package payment

type PaymentMethodResponse struct {
	ID       int     `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	FeeType  FeeType `json:"feeType"`
	FeeValue float64 `json:"feeValue"`
	IsActive bool    `json:"isActive"`
}

func ToPaymentMethodResponse(pm PaymentMethods) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:       pm.ID,
		Code:     pm.Code,
		Name:     pm.Name,
		FeeType:  pm.FeeType,
		FeeValue: pm.FeeValue,
		IsActive: pm.IsActive,
	}
}

func ToPaymentMethodListResponse(pms []PaymentMethods) []PaymentMethodResponse {
	list := make([]PaymentMethodResponse, len(pms))
	for i, pm := range pms {
		list[i] = ToPaymentMethodResponse(pm)
	}
	return list
}
