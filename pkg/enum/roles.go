package enum

type RoleName string

const (
	RoleUser               RoleName = "user"
	RoleChairman           RoleName = "chairman"
	RoleSocialManager      RoleName = "social_manager"
	RoleFinance            RoleName = "finance"
	RoleAmbulanceManager   RoleName = "ambulance_manager"
	RoleAmbulanceDriver    RoleName = "ambulance_driver"
	RolePublicationManager RoleName = "publication_manager"
	RoleSuperadmin         RoleName = "superadmin"
)
