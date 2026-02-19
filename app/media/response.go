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
