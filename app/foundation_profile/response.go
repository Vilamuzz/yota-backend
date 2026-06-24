package foundation_profile

import (
	"time"

	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type FoundationProfileResponse struct {
	ID                    string    `json:"id"`
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

func (f *FoundationProfile) toFoundationProfileResponse() FoundationProfileResponse {
	return FoundationProfileResponse{
		ID:                    f.ID.String(),
		FoundationName:        f.FoundationName,
		FounderPicture:        s3_pkg.GetCDNURL(f.FounderPicture),
		FounderName:           f.FounderName,
		FoundationAddress:     f.FoundationAddress,
		FoundationPhone:       f.FoundationPhone,
		FoundationEmail:       f.FoundationEmail,
		FoundationInstagram:   f.FoundationInstagram,
		FoundationFacebook:    f.FoundationFacebook,
		FoundationTwitter:     f.FoundationTwitter,
		EmbeddedAddress:       f.EmbeddedAddress,
		Logo:                  s3_pkg.GetCDNURL(f.Logo),
		Icon:                  s3_pkg.GetCDNURL(f.Icon),
		OrganizationStructure: s3_pkg.GetCDNURL(f.OrganizationStructure),
		HeroImageOne:          s3_pkg.GetCDNURL(f.HeroImageOne),
		HeroImageTwo:          s3_pkg.GetCDNURL(f.HeroImageTwo),
		HeroImageThree:        s3_pkg.GetCDNURL(f.HeroImageThree),
		HeroImageFour:         s3_pkg.GetCDNURL(f.HeroImageFour),
		CreatedAt:             f.CreatedAt,
		UpdatedAt:             f.UpdatedAt,
	}
}
