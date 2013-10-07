package util

func InSlice(element int, array []int) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	
	return false
}
