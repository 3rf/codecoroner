package unused

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUnusedIdentsWithMain(t *testing.T) {
	Convey("with a test main package and a default UnusedCodeFinder", t, func() {
		ucf := NewUnusedCodeFinder()
		So(ucf, ShouldNotBeNil)
		ucf.Idents = true

		Convey("running 'idents'", func() {
			results, err := ucf.Run([]string{"testdata"})
			So(err, ShouldBeNil)

			Convey("all idents in pkg1, pkg2, and main should be found", func() {
				So("pkg1.Number", ShouldBeFoundIn, results)
				So("pkg1.AnotherNumber", ShouldBeFoundIn, results)
				So("pkg1.GenSix", ShouldBeFoundIn, results)
				So("pkg1.GenUInt", ShouldBeFoundIn, results)
				So("pkg2.(unusedType).Val", ShouldBeFoundIn, results)
				So("pkg2.GrayKittenLink", ShouldBeFoundIn, results)
				So("oldHelper", ShouldBeFoundIn, results)
				So("unusedParam", ShouldBeFoundIn, results)

				So("GenInt", ShouldNotBeFoundIn, results)
				So("GenIntMod400", ShouldNotBeFoundIn, results)
				So("ColorKittenLink", ShouldNotBeFoundIn, results)
				So("init", ShouldNotBeFoundIn, results)
			})

			Convey("but funcs that are called in other unused funcs will not be found", func() {
				So("toUint", ShouldNotBeFoundIn, results)
			})
		})
	})
}

func TestUnusedIdentsWithTests(t *testing.T) {
	Convey("with a test main package and a default UnusedCodeFinder", t, func() {
		ucf := NewUnusedCodeFinder()
		So(ucf, ShouldNotBeNil)
		ucf.Idents = true
		ucf.IncludeTests = true

		Convey("running 'idents' with -tests", func() {
			results, err := ucf.Run([]string{"testdata"})
			So(err, ShouldBeNil)

			Convey("all idents in pkg1, pkg2, and main should be found", func() {
				So("pkg1.Number", ShouldBeFoundIn, results)
				So("pkg1.AnotherNumber", ShouldBeFoundIn, results)
				So("pkg1.GenUInt", ShouldBeFoundIn, results)
				So("pkg2.(unusedType).Val", ShouldBeFoundIn, results)
				So("pkg2.GrayKittenLink", ShouldBeFoundIn, results)
				So("oldHelper", ShouldBeFoundIn, results)
				So("unusedParam", ShouldBeFoundIn, results)

				So("GenInt", ShouldNotBeFoundIn, results)
				So("GenIntMod400", ShouldNotBeFoundIn, results)
				So("ColorKittenLink", ShouldNotBeFoundIn, results)
				So("init", ShouldNotBeFoundIn, results)
				So("toUint", ShouldNotBeFoundIn, results)

				Convey("plus idents only found in tests", func() {
					So("pkg1.testhelper", ShouldBeFoundIn, results)
					So("pkg1.GenSix", ShouldNotBeFoundIn, results)
				})
			})
		})
	})
}
