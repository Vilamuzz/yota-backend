package media

type MediaRequest struct {
	ID  string `json:"id" binding:"omitempty,uuid"`
	URL string `json:"url" binding:"required,url"`

	Type    string `json:"type" binding:"required,oneof=image video"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order" binding:"omitempty,min=0"`
}
