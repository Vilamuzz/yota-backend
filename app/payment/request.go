package payment

type UpdatePaymentMethodRequest struct {
	FeeType  FeeType `json:"feeType" binding:"required,oneof=flat percentage"`
	FeeValue float64 `json:"feeValue" binding:"min=0"`
	IsActive *bool   `json:"isActive" binding:"required"`
}
