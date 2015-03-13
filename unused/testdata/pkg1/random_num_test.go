package pkg1

import (
	"testing"
)

// This helper should be found by [idents] and [funcs] if
// test analysis is enabled.
func testhelper() int {
	return 7
}

// This should not be found
func TestTheNumberSix(t *testing.T) {
	if GenSix() != 6 {
		t.Fatal("THIS HAS GONE POORLY")
	}
}
