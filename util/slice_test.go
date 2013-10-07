package util

import (
	"testing"
)

func TestInArray(t *testing.T) {
	fixtures := []int {3, 5, 10}
	if !InSlice(3, fixtures) {
		t.Error("3 should be in slice")
	}
}
