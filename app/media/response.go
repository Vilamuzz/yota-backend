package media

type PublishedMediaResponse struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order"`
}

type MediaResponse struct {
	ID         string `json:"id"`
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	AltText    string `json:"alt_text"`
	Order      int    `json:"order"`
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
		ID:         m.ID,
		EntityID:   m.EntityID,
		EntityType: m.EntityType,
		Type:       m.Type,
		URL:        m.URL,
		AltText:    m.AltText,
		Order:      m.Order,
	}
}
