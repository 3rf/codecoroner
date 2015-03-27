package unused

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func init() {
	// quick check to make sure we are in the right directory
	if _, err := os.Stat("testdata/mockmain.go"); err != nil {
		panic("unused tests must be run from the 'github.com/3rf/codecoroner/unused' dir")
	}
}

// helpers for convey
func ShouldBeFoundIn(actual interface{}, expected ...interface{}) string {
	// this can panic, but I'm not adding type checking
	target := actual.(string)
	results := expected[0].([]UnusedThing)
	for _, thing := range results {
		if thing.Name == target {
			return ""
		}
	}
	return fmt.Sprintf("nothing named '%v' found in results", target)
}

func ShouldNotBeFoundIn(actual interface{}, expected ...interface{}) string {
	// this can panic, but I'm not adding type checking
	target := actual.(string)
	results := expected[0].([]UnusedThing)
	for _, thing := range results {
		if thing.Name == target {
			return fmt.Sprintf("found '%v' in results (it shouldn't be there)", target)
		}
	}
	return ""
}

func TestUnusedFuncsWithMain(t *testing.T) {
	Convey("with a test main package and a default UnusedCodeFinder", t, func() {
		ucf := NewUnusedCodeFinder()
		So(ucf, ShouldNotBeNil)

		Convey("running 'funcs'", func() {
			results, err := ucf.Run([]string{"testdata"})
			So(err, ShouldBeNil)

			Convey("all functions in pkg1 and pkg2 that main does not use should be found", func() {
				So("oldHelper", ShouldBeFoundIn, results)
				So("GenSix", ShouldBeFoundIn, results)
				So("GenUInt", ShouldBeFoundIn, results)
				So("toUint", ShouldBeFoundIn, results)
				So("GrayKittenLink", ShouldBeFoundIn, results)
				So("GenInt", ShouldNotBeFoundIn, results)
				So("GenIntMod400", ShouldNotBeFoundIn, results)
				So("ColorKittenLink", ShouldNotBeFoundIn, results)
			})
		})
	})
}

func TestUnusedFuncsWithAll(t *testing.T) {
	Convey("with a test main package and a UnusedCodeFinder with -all", t, func() {
		ucf := NewUnusedCodeFinder()
		So(ucf, ShouldNotBeNil)
		ucf.IncludeAll = true

		Convey("running 'funcs'", func() {
			results, err := ucf.Run([]string{"testdata"})
			So(err, ShouldBeNil)

			Convey("all functions that are unused by any pkg are found", func() {
				So("oldHelper", ShouldBeFoundIn, results)
				So("GenUInt", ShouldBeFoundIn, results)
				So("toUint", ShouldBeFoundIn, results)
				So("GrayKittenLink", ShouldBeFoundIn, results)
			})

			Convey("but GenSix should not be found, since it is used in a test", func() {
				So("GenSix", ShouldNotBeFoundIn, results)
			})
		})
	})
}
