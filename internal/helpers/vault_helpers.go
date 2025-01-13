package helpers

var DisabledTiers = []string{"STARTER_SMALL"}

func IsDisabledTier(v string) bool {
	for _, tier := range DisabledTiers {
		if tier == v {
			return true
		}
	}
	return false
}
