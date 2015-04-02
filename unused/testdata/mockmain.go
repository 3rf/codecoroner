package main

import (
	"fmt"
	"github.com/3rf/codecoroner/unused/testdata/pkg1"
	"github.com/3rf/codecoroner/unused/testdata/pkg2"
)

func init() {
	// do nothing
}

// this function should be found by both modes,
// the "unusedParam" parameter should be found by [idents]
func oldHelper(str string, unusedParam uintptr) int {
	return len(str)
}

func main() {
	fmt.Println("This program is just for testing codecoroner.")
	fmt.Println("You're welcome to just run it if you want,")
	fmt.Println("but it's not meant to actually be used...\n")

	fmt.Println("Here are some random numbers:", pkg1.GenInt(), pkg1.GenInt(), pkg1.GenInt())
	fmt.Println("And here is a link to a picture of a cat:", pkg2.ColorKittenLink())
}
