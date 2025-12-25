package tenant

type tenantTier string

const (
	tenantTierFree  tenantTier = "free"
	tenantTierBasic            = "basic"
	tenantTierPro              = "pro"
)

type tenantRegion string

const (
	usEast1 tenantRegion = "us-east-1"
	usEast2 tenantRegion = "us-east-2"
)

func isValidRegion(region string) bool {
	switch region {
	case string(usEast1), string(usEast2):
		return true
	default:
		return false
	}
}
