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
				So("Number", ShouldBeFoundIn, results)
				So("AnotherNumber", ShouldBeFoundIn, results)
				So("GenSix", ShouldBeFoundIn, results)
				So("GenUInt", ShouldBeFoundIn, results)
				So("(unusedType).Val", ShouldBeFoundIn, results)
				So("field", ShouldBeFoundIn, results)
				So("GrayKittenLink", ShouldBeFoundIn, results)
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
				So("Number", ShouldBeFoundIn, results)
				So("AnotherNumber", ShouldBeFoundIn, results)
				So("GenUInt", ShouldBeFoundIn, results)
				So("(unusedType).Val", ShouldBeFoundIn, results)
				So("field", ShouldBeFoundIn, results)
				So("GrayKittenLink", ShouldBeFoundIn, results)
				So("oldHelper", ShouldBeFoundIn, results)
				So("unusedParam", ShouldBeFoundIn, results)

				So("GenInt", ShouldNotBeFoundIn, results)
				So("GenIntMod400", ShouldNotBeFoundIn, results)
				So("ColorKittenLink", ShouldNotBeFoundIn, results)
				So("init", ShouldNotBeFoundIn, results)
				So("toUint", ShouldNotBeFoundIn, results)

				Convey("plus idents only found in tests", func() {
					So("testhelper", ShouldBeFoundIn, results)
					So("GenSix", ShouldNotBeFoundIn, results)
				})
			})
		})
	})
}

func TestUnusedIdentsWithIgnore(t *testing.T) {
	Convey("with a test main package and a default UnusedCodeFinder", t, func() {
		ucf := NewUnusedCodeFinder()
		So(ucf, ShouldNotBeNil)
		ucf.Idents = true
		ucf.Ignore = []string{"pkg1", "pkg2"}

		Convey("running 'idents' with -ignore to skip pkg1 and pk2", func() {
			results, err := ucf.Run([]string{"testdata"})
			So(err, ShouldBeNil)

			Convey("only unused idents in main should be found", func() {
				So("oldHelper", ShouldBeFoundIn, results)
				So("unusedParam", ShouldBeFoundIn, results)

				Convey("and nothing else", func() {
					So("pkg1.Number", ShouldNotBeFoundIn, results)
					So("pkg1.AnotherNumber", ShouldNotBeFoundIn, results)
					So("pkg1.GenUInt", ShouldNotBeFoundIn, results)
					So("pkg2.GrayKittenLink", ShouldNotBeFoundIn, results)
				})
			})
		})
	})
}
