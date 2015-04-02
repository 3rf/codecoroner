// A small test package that generates random kitten picture links.
// This code is test code for codecoroner analysis--do not actually use it.
package pkg2

import (
	"fmt"
	"github.com/3rf/codecoroner/unused/testdata/pkg1"
)

// this type and its method should be found by [idents]
type unusedType struct{ field int }

func (ut unusedType) Val() int {
	return 2
}

// This function should not be found, as it is used.
func ColorKittenLink() string {
	return fmt.Sprintf("http://placekitten.com/%v/%v",
		pkg1.GenIntMod400()+400,
		pkg1.GenIntMod400()+200)
}

// This function should be found by [idents] and [funcs]
func GrayKittenLink() string {
	return fmt.Sprintf("http://placekitten.com/g/%v/%v",
		pkg1.GenIntMod400()+400,
		pkg1.GenIntMod400()+200)
}
