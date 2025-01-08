package helpers

var DISABLED_TIERS = []string{"STARTER_SMALL"}

func IsDisabledTier(v string) bool {
	for _, tier := range DISABLED_TIERS {
		if tier == v {
			return true
		}
	}
	return false
}
