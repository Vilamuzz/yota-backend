package image

type Image struct {
	ID         string `json:"id" gorm:"primary_key"`
	EntityID   string `json:"entity_id" gorm:"not null"`
	EntityType string `json:"entity_type" gorm:"not null"`
	URL        string `json:"url" gorm:"not null"`
	AltText    string `json:"alt_text"`
	Order      int    `json:"order" gorm:"not null;default:0"`
}
