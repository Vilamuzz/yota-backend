package payment

type FeeType string

const (
	FeeTypeFlat       FeeType = "flat"
	FeeTypePercentage FeeType = "percentage"
)

type PaymentMethods struct {
	ID       int     `json:"id" gorm:"primaryKey"`
	Code     string  `json:"code" gorm:"not null"`
	Name     string  `json:"name" gorm:"not null"`
	FeeType  FeeType `json:"feeType" gorm:"not null"`
	FeeValue float64 `json:"feeValue" gorm:"not null"`
	IsActive bool    `json:"isActive" gorm:"not null"`
}
