package main

import (
	"fmt"
	"sync"
	"time"
)

var pkgVar = 123

type PkgType1 struct {
	myInt    int
	myStr    string
	internal struct {
		myByte byte
	}
}

type PkgType2 struct {
	time.Time
}

type PkgType3 struct {
	*sync.Mutex
}

var PkgAnonStruct struct {
	field1 time.Time
}

type funcType func(typeParam int) error

func ReturnOne(ignoreParam int) int {
	type internalType struct {
		myInt     int64
		myFloat64 float64
	}
	a := internalType{}
	a.myInt = 1
	return int(a.myInt)
}

type printer struct{}

func (p printer) Print(a interface{}) {
	fmt.Println(a)
}

func main() {
	printer{}.Print(ReturnOne(777))

	doNothing := func(innerIgnore int) {
		return
	}

	localAnonStruct := struct {
		field2 int
	}{}
	_ = localAnonStruct

	go func(anonParam int) {
		doNothing(4)
	}(123)

}
