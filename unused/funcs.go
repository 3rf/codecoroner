package unused

import (
	"fmt"
	"go/build"
	"golang.org/x/tools/oracle"
	"strings"
)

func (ucf *UnusedCodeFinder) getCallgraphFromOracle() error {
	res, err := oracle.Query(ucf.pkgsAsArray(), "callgraph", "", nil, &build.Default, true)
	if err != nil {
		return err
	}
	serialRes := res.Serial()
	if serialRes.Callgraph == nil {
		return fmt.Errorf("no callgraph present in oracle results")
	}
	ucf.Callgraph = serialRes.Callgraph
	return nil
}

func (ucf *UnusedCodeFinder) isInCG(f UnusedThing) bool {
	files, ok := ucf.filesByCaller[f.Name]
	if !ok {
		return false
	}
	for _, path := range files {
		if strings.Contains(path, f.File) {
			return true
		}
	}
	return false
}

func (ucf *UnusedCodeFinder) computeUnusedFuncs() []UnusedThing {
	unused := []UnusedThing{}
	for _, f := range ucf.funcs {
		if !ucf.isInCG(f) {
			unused = append(unused, f)
		}
	}
	return unused
}

func (ucf *UnusedCodeFinder) buildFileMap() {
	for _, entry := range ucf.Callgraph {
		//strip off the package name for simplicity
		//TODO, can this be left on? Try prepending func names with package?
		idx := strings.LastIndex(entry.Name, ".") + 1
		if idx != 0 {
			ucf.filesByCaller[entry.Name[idx:]] = append(ucf.filesByCaller[entry.Name[idx:]], entry.Pos)
		}
	}
}
