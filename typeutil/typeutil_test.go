package typeutil

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"go/ast"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"testing"
)

const testMainPkg = "github.com/3rf/codecoroner/typeutil/testdata"

func loadMainInfo() *loader.PackageInfo {
	var conf loader.Config
	_, err := conf.FromArgs([]string{testMainPkg}, false)
	if err != nil {
		panic(fmt.Sprintf("error loading test program data: %v", err))
	}
	p, err := conf.Load()
	if err != nil {
		panic(fmt.Sprintf("error loading program data: %v", err))
	}
	if !p.Imported[testMainPkg].TransitivelyErrorFree {
		panic(fmt.Sprintf("error loading program"))
	}
	return p.Imported[testMainPkg]
}

func findObjectWithName(name string, objs map[*ast.Ident]types.Object) types.Object {
	for key, val := range objs {
		if key.Name == name {
			return val
		}
	}
	return nil
}

func TestLookupFuncForParameter(t *testing.T) {
	Convey("with a test main package", t, func() {
		info := loadMainInfo()

		Convey("and the types.Object for var ignoreParam", func() {
			ignoreParam := findObjectWithName("ignoreParam", info.Defs)
			So(ignoreParam, ShouldNotBeNil)

			Convey("running LookupFunctionForParameter should return ReturnOne", func() {
				f := LookupFuncForParameter(ignoreParam.(*types.Var))
				So(f, ShouldNotBeNil)
				So(f.Name(), ShouldEqual, "ReturnOne")
			})
		})

		Convey("and the types.Object for var innerIgnore", func() {
			innerIgnore := findObjectWithName("innerIgnore", info.Defs)
			So(innerIgnore, ShouldNotBeNil)

			Convey("running LookupFunctionForParameter should return ReturnOne", func() {
				f := LookupFuncForParameter(innerIgnore.(*types.Var))
				So(f, ShouldNotBeNil)
				So(f.Name(), ShouldEqual, "doNothing")
			})
		})

		Convey("and the types.Object for var anonParam from an anonymous func", func() {
			anonParam := findObjectWithName("anonParam", info.Defs)
			So(anonParam, ShouldNotBeNil)

			Convey("running LookupFunctionForParameter should return nil", func() {
				f := LookupFuncForParameter(anonParam.(*types.Var))
				So(f, ShouldBeNil)
			})
		})
	})
}

func TestLookupStructForField(t *testing.T) {
	Convey("with a test main package", t, func() {
		info := loadMainInfo()

		Convey("and the types.Object for field (PkgType1).myStr", func() {
			myStr := findObjectWithName("myStr", info.Defs)
			So(myStr, ShouldNotBeNil)

			Convey("running LookupStructForField should return PkgType1", func() {
				s := LookupStructForField(myStr.(*types.Var))
				So(s, ShouldNotBeNil)
				So(s.Name(), ShouldEqual, "PkgType1")
			})
		})

		Convey("and the types.Object for the nested field (PkgType1).myByte", func() {
			myByte := findObjectWithName("myByte", info.Defs)
			So(myByte, ShouldNotBeNil)

			Convey("running LookupStructForField should return PkgType1", func() {
				s := LookupStructForField(myByte.(*types.Var))
				So(s, ShouldNotBeNil)
				So(s.Name(), ShouldEqual, "PkgType1")
			})
		})

		Convey("and the types.Object for field (internalType).myFloat64", func() {
			myFloat64 := findObjectWithName("myFloat64", info.Defs)
			So(myFloat64, ShouldNotBeNil)

			Convey("running LookupStructForField should return internalType", func() {
				s := LookupStructForField(myFloat64.(*types.Var))
				So(s, ShouldNotBeNil)
				So(s.Name(), ShouldEqual, "internalType")
			})
		})

		Convey("with different fields both named myInt", func() {
			myInt := findObjectWithName("myInt", info.Uses)
			So(myInt, ShouldNotBeNil)

			Convey("running LookupStructForField should return the correct struct", func() {
				s := LookupStructForField(myInt.(*types.Var))
				So(s, ShouldNotBeNil)
				So(s.Name(), ShouldEqual, "internalType")
			})
		})

		Convey("and a types.Object for a var that ISN'T a field", func() {
			pkgVar := findObjectWithName("pkgVar", info.Defs)
			So(pkgVar, ShouldNotBeNil)

			Convey("running LookupStructForField should return nil", func() {
				s := LookupStructForField(pkgVar.(*types.Var))
				So(s, ShouldBeNil)
			})
		})
	})
}
