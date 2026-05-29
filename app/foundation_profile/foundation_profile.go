package foundation_profile

import (
	"time"

	"github.com/google/uuid"
)

type FoundationProfile struct {
	ID                    uuid.UUID `json:"id" gorm:"primaryKey"`
	FoundationName        string    `json:"foundationName"`
	FounderPicture        string    `json:"founderPicture"`
	FounderName           string    `json:"founderName"`
	FoundationAddress     string    `json:"foundationAddress"`
	FoundationPhone       string    `json:"foundationPhone"`
	FoundationEmail       string    `json:"foundationEmail"`
	FoundationInstagram   *string   `json:"foundationInstagram"`
	FoundationFacebook    *string   `json:"foundationFacebook"`
	FoundationTwitter     *string   `json:"foundationTwitter"`
	EmbeddedAddress       string    `json:"embeddedAddress"`
	Logo                  string    `json:"logo"`
	Icon                  string    `json:"icon"`
	OrganizationStructure string    `json:"organizationStructure"`
	HeroImageOne          string    `json:"heroImageOne"`
	HeroImageTwo          string    `json:"heroImageTwo"`
	HeroImageThree        string    `json:"heroImageThree"`
	HeroImageFour         string    `json:"heroImageFour"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
