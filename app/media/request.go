package media

type MediaItem struct {
	URL     string `json:"url" binding:"required,url"`
	Type    string `json:"type" binding:"required,oneof=image video"`
	AltText string `json:"alt_text"`
}
