package media

type MediaRequest struct {
	ID      string `json:"id" binding:"omitempty,uuid"`
	URL     string `json:"url" binding:"required,url"`
	Action  string `json:"action" gorm:"not null"` // create, update, delete
	Type    string `json:"type" binding:"required,oneof=image video"`
	AltText string `json:"alt_text"`
}
