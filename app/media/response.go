package media

type PublishedMediaResponse struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order"`
}

type MediaResponse struct {
	ID        string `json:"id"`
	NewsID    string `json:"news_id"`
	GalleryID string `json:"gallery_id"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	Order     int    `json:"order"`
}

func (m *Media) toPublishedMediaResponse() PublishedMediaResponse {
	return PublishedMediaResponse{
		URL:     m.URL,
		AltText: m.AltText,
		Order:   m.Order,
	}
}

func (m *Media) toMediaResponse() MediaResponse {
	return MediaResponse{
		ID:        m.ID.String(),
		NewsID:    m.NewsID.String(),
		GalleryID: m.GalleryID.String(),
		Type:      string(m.Type),
		URL:       m.URL,
		AltText:   m.AltText,
		Order:     m.Order,
	}
}
