package typeutils

import (
	"fmt"
	"go/ast"
	"go/types"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/tools/go/loader"
)

const testMainPkg = "github.com/3rf/codecoroner/typeutils/testdata"

func loadMainInfo() *loader.PackageInfo {
	p := loadProg()
	return p.Imported[testMainPkg]
}

func loadProg() *loader.Program {
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
	return p
}

func findObjectWithName(name string, objs map[*ast.Ident]types.Object) types.Object {
	for key, val := range objs {
		if key.Name == name {
			return val
		}
	}
	return nil
}

func TestParameterMethods(t *testing.T) {
	Convey("with a test main package", t, func() {
		info := loadMainInfo()
		prog := Program(loadProg())

		Convey("and the types.Object for var ignoreParam", func() {
			ignoreParam := findObjectWithName("ignoreParam", info.Defs)
			So(ignoreParam, ShouldNotBeNil)

			Convey("running IsParameter should return true", func() {
				So(prog.IsParameter(ignoreParam.(*types.Var)), ShouldBeTrue)
			})

			Convey("running LookupFunctionForParameter should return ReturnOne", func() {
				f := LookupFuncForParameter(ignoreParam.(*types.Var))
				So(f, ShouldNotBeNil)
				So(f.Name(), ShouldEqual, "ReturnOne")
				prog.FunctionForParameter(ignoreParam.(*types.Var))
			})
		})

		Convey("and the types.Object for var innerIgnore", func() {
			innerIgnore := findObjectWithName("innerIgnore", info.Defs)
			So(innerIgnore, ShouldNotBeNil)

			Convey("running IsParameter should return true", func() {
				So(prog.IsParameter(innerIgnore.(*types.Var)), ShouldBeTrue)
			})

			Convey("running LookupFunctionForParameter should return doNothing", func() {
				f := LookupFuncForParameter(innerIgnore.(*types.Var))
				So(f, ShouldNotBeNil)
				So(f.Name(), ShouldEqual, "doNothing")
				prog.FunctionForParameter(innerIgnore.(*types.Var))
			})
		})

		Convey("and the types.Object for var anonParam from an anonymous func", func() {
			anonParam := findObjectWithName("anonParam", info.Defs)
			So(anonParam, ShouldNotBeNil)

			Convey("running IsParameter should return true", func() {
				So(prog.IsParameter(anonParam.(*types.Var)), ShouldBeTrue)
			})

			Convey("running LookupFunctionForParameter should return nil", func() {
				f := LookupFuncForParameter(anonParam.(*types.Var))
				So(f, ShouldBeNil)
				prog.FunctionForParameter(anonParam.(*types.Var))
			})
		})

		Convey("and the types.Object for var doNothing, which ISN'T a param", func() {
			doNothing := findObjectWithName("doNothing", info.Defs)
			So(doNothing, ShouldNotBeNil)

			Convey("running IsParameter should return true", func() {
				So(prog.IsParameter(doNothing.(*types.Var)), ShouldBeFalse)
			})
		})
	})
}

func TestMethodMethods(t *testing.T) {
	Convey("with a test main package", t, func() {
		info := loadMainInfo()
		prog := Program(loadProg())

		Convey("and the types.Object for method printer.Print", func() {
			fPrint := findObjectWithName("Print", info.Defs)
			So(fPrint, ShouldNotBeNil)

			Convey("running IsMethod should return true", func() {
				So(prog.IsMethod(fPrint.(*types.Func)), ShouldBeTrue)
			})
		})

		Convey("and the types.Object for func ReturnOne, which ISN'T a method", func() {
			ReturnOne := findObjectWithName("ReturnOne", info.Defs)
			So(ReturnOne, ShouldNotBeNil)

			Convey("running IsMethod should return false", func() {
				So(prog.IsMethod(ReturnOne.(*types.Func)), ShouldBeFalse)
			})
		})

	})
}

func TestStructForField(t *testing.T) {
	Convey("with a test main package", t, func() {
		info := loadMainInfo()
		prog := Program(loadProg())

		Convey("and the types.Object for field (PkgType1).myStr", func() {
			myStr := findObjectWithName("myStr", info.Defs)
			So(myStr, ShouldNotBeNil)

			Convey("running IsStructField should return true", func() {
				So(prog.IsStructField(myStr.(*types.Var)), ShouldBeTrue)
			})

			Convey("running StructForField should return PkgType1", func() {
				s := prog.StructForField(myStr.(*types.Var))
				So(s, ShouldEqual, "PkgType1")
			})
		})

		Convey("and the types.Object for the nested field (PkgType1).myByte", func() {
			myByte := findObjectWithName("myByte", info.Defs)
			So(myByte, ShouldNotBeNil)

			Convey("running IsStructField should return true", func() {
				So(prog.IsStructField(myByte.(*types.Var)), ShouldBeTrue)
			})

			Convey("running StructForField should return PkgType1", func() {
				s := prog.StructForField(myByte.(*types.Var))
				So(s, ShouldEqual, "PkgType1.internal")
			})
		})

		Convey("and the types.Object for field (internalType).myFloat64", func() {
			myFloat64 := findObjectWithName("myFloat64", info.Defs)
			So(myFloat64, ShouldNotBeNil)

			Convey("running IsStructField should return true", func() {
				So(prog.IsStructField(myFloat64.(*types.Var)), ShouldBeTrue)
			})

			Convey("running StructForField should return internalType", func() {
				s := prog.StructForField(myFloat64.(*types.Var))
				So(s, ShouldEqual, "internalType")
			})
		})

		Convey("with different fields both named myInt", func() {
			myInt := findObjectWithName("myInt", info.Uses)
			So(myInt, ShouldNotBeNil)

			Convey("running StructForField should return internalType", func() {
				s := prog.StructForField(myInt.(*types.Var))
				So(s, ShouldEqual, "internalType")
			})
		})

		Convey("and a types.Object for a var that ISN'T a field", func() {
			pkgVar := findObjectWithName("pkgVar", info.Defs)
			So(pkgVar, ShouldNotBeNil)

			Convey("running IsStructField should return false", func() {
				So(prog.IsStructField(pkgVar.(*types.Var)), ShouldBeFalse)
			})

			Convey("running StructForField should return nothing", func() {
				s := prog.StructForField(pkgVar.(*types.Var))
				So(s, ShouldEqual, "")
			})
		})

		Convey("with a types.Object for an anonymous package struct field", func() {
			anon := findObjectWithName("field1", info.Defs)
			So(anon, ShouldNotBeNil)

			Convey("running IsStructField should return true", func() {
				So(prog.IsStructField(anon.(*types.Var)), ShouldBeTrue)
			})

			Convey("StructForField should be the name of the variable", func() {
				So(prog.StructForField(anon.(*types.Var)), ShouldEqual, "PkgAnonStruct")
			})
		})

		Convey("with a types.Object for an anonymous local struct field", func() {
			anon := findObjectWithName("field2", info.Defs)
			So(anon, ShouldNotBeNil)

			Convey("running IsStructField should return true", func() {
				So(prog.IsStructField(anon.(*types.Var)), ShouldBeTrue)
			})

			Convey("StructForField should be the name of the variable", func() {
				So(prog.StructForField(anon.(*types.Var)), ShouldEqual, "localAnonStruct")
			})
		})
	})
}
