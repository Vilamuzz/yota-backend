package foundation_profile

import "mime/multipart"

type FoundationProfileCreateRequest struct {
	FoundationName        string                `form:"foundationName" json:"foundationName"`
	FounderPicture        *multipart.FileHeader `form:"founderPicture" swaggerignore:"true"`
	FounderName           string                `form:"founderName" json:"founderName"`
	FoundationAddress     string                `form:"foundationAddress" json:"foundationAddress"`
	FoundationPhone       string                `form:"foundationPhone" json:"foundationPhone"`
	FoundationEmail       string                `form:"foundationEmail" json:"foundationEmail"`
	FoundationInstagram   *string               `form:"foundationInstagram" json:"foundationInstagram"`
	FoundationFacebook    *string               `form:"foundationFacebook" json:"foundationFacebook"`
	FoundationTwitter     *string               `form:"foundationTwitter" json:"foundationTwitter"`
	EmbeddedAddress       string                `form:"embeddedAddress" json:"embeddedAddress"`
	Logo                  *multipart.FileHeader `form:"logo" swaggerignore:"true"`
	Icon                  *multipart.FileHeader `form:"icon" swaggerignore:"true"`
	OrganizationStructure *multipart.FileHeader `form:"organizationStructure" swaggerignore:"true"`
	HeroImageOne          *multipart.FileHeader `form:"heroImageOne" swaggerignore:"true"`
	HeroImageTwo          *multipart.FileHeader `form:"heroImageTwo" swaggerignore:"true"`
	HeroImageThree        *multipart.FileHeader `form:"heroImageThree" swaggerignore:"true"`
	HeroImageFour         *multipart.FileHeader `form:"heroImageFour" swaggerignore:"true"`
}

type FoundationProfileUpdateRequest struct {
	FoundationName        string                `form:"foundationName" json:"foundationName"`
	FounderPicture        *multipart.FileHeader `form:"founderPicture" swaggerignore:"true"`
	FounderName           string                `form:"founderName" json:"founderName"`
	FoundationAddress     string                `form:"foundationAddress" json:"foundationAddress"`
	FoundationPhone       string                `form:"foundationPhone" json:"foundationPhone"`
	FoundationEmail       string                `form:"foundationEmail" json:"foundationEmail"`
	FoundationInstagram   *string               `form:"foundationInstagram" json:"foundationInstagram"`
	FoundationFacebook    *string               `form:"foundationFacebook" json:"foundationFacebook"`
	FoundationTwitter     *string               `form:"foundationTwitter" json:"foundationTwitter"`
	EmbeddedAddress       string                `form:"embeddedAddress" json:"embeddedAddress"`
	Logo                  *multipart.FileHeader `form:"logo" swaggerignore:"true"`
	Icon                  *multipart.FileHeader `form:"icon" swaggerignore:"true"`
	OrganizationStructure *multipart.FileHeader `form:"organizationStructure" swaggerignore:"true"`
	HeroImageOne          *multipart.FileHeader `form:"heroImageOne" swaggerignore:"true"`
	HeroImageTwo          *multipart.FileHeader `form:"heroImageTwo" swaggerignore:"true"`
	HeroImageThree        *multipart.FileHeader `form:"heroImageThree" swaggerignore:"true"`
	HeroImageFour         *multipart.FileHeader `form:"heroImageFour" swaggerignore:"true"`
}
