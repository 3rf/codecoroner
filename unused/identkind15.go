// +build go1.5

package unused

import "go/types"

func objToFunc(obj types.Object) (f typeFunc, ok bool) {
	f, ok = obj.(*types.Func)
	return f, ok
}
