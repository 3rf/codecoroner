package pkg1

import (
	"testing"
)

// This should never be found, regardless of tests setting
func testhelper() int {
	return 7
}

// This should only be found with tests disabled
func TestTheNumberSix(t *testing.T) {
	if GenSix() != 6 {
		t.Fatal("THIS HAS GONE POORLY")
	}
}
