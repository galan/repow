package util

func IsInSlice(value string, list ...string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
