package unused

// typeFunc is implemented by both go/types.Func
// and golang.org/x/tools/go/types/Func
type typeFunc interface {
	Name() string
	FullName() string
}
