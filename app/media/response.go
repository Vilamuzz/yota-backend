package media

type PublishedMediaResponse struct {
	URL   string `json:"url"`
	Alt   string `json:"alt"`
	Order int    `json:"order"`
}

type MediaResponse struct {
	ID        string `json:"id"`
	NewsID    string `json:"newsId"`
	GalleryID string `json:"galleryId"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	Order     int    `json:"order"`
}
