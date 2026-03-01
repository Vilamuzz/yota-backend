package pray

type Pray struct {
	ID string `json:"id" gorm:"primary_key"`
	DonationID string `json:"donation_id" gorm:"not null"`
	UserID string `json:"user_id" gorm:"not null"`
	
}