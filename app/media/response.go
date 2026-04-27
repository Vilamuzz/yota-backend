package media

type PublishedMediaResponse struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order"`
}

type MediaResponse struct {
	ID        string `json:"id"`
	NewsID    string `json:"newsId"`
	GalleryID string `json:"galleryId"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	AltText   string `json:"altText"`
	Order     int    `json:"order"`
}
