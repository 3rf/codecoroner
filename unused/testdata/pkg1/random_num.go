// A small test package that generates random numbers. This code
// is test code for codecoroner analysis--do not actually use it.
package pkg1

import (
	"math/rand"
)

// This const should be found by [idents]
const Number = 5

// This var should be found by [idents]
var AnotherNumber = 7

// This should not be found by any mode.
var Six = 6

// This function is used, so it should not be found by any mode.
func GenInt() int {
	return rand.Int()
}

// This function should only be found by [idents] and [funcs] if
// pkg2 is left out (i.e. just pkg1 is analyzed).
func GenIntMod400() int {
	return GenInt() % 400
}

// This function should be found by [funcs] but not [idents], since
// it is called by GenUInt, which is a dead function.
func toUint(i int) uint {
	return uint(i)
}

// This function isn't used by any package, so [funcs] should find it
func GenUInt() uint {
	return toUint(GenInt())
}

// This function is only used in testing, so it should only be found
// when tests are not included.
func GenSix() int {
	return Six
}
