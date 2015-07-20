codecoroner [![Build Status](https://travis-ci.org/3rf/codecoroner.svg)](https://travis-ci.org/3rf/codecoroner)
===============
######Version 1.1 by Kyle Erf, MIT License 


Leaving dead code in a large codebase with multiple libraries is difficult to avoid.
Things get moved around; functions get refactored, leaving helpers on their own; people miscommunicate. 

One of the easiest ways to detect dead code is through static analysis. 
Unfortunately, Go's current static analysis tools (`oracle`, `callgraph`, etc) do not make aggregation of unused functions as easy as it should be.
This tool, codecoroner, uses the output of the Go `ssa`, `callgraph`, and `types` libraries to find unused functions/methods in your codebase. 

###Quick Start

First, grab the go-unused-funcs binary by either cloning this git repository and building main.go or by running
```bash
go get github.com/3rf/codecoroner
```
which should install a go-unused-funcs binary in `$GOPATH/bin`

Codecoroner takes in a set of files/folders, and scans them for functions and packages.
Every "main" package is collected and given to the go oracle for callgraph analysis.
The tool then checks all of the scanned function declarations against the callgraph to find unused ones.
will check the go-unused-funcs source folder for unused functions.

Notes:
 * Run with `-v` to print a log of what go-unused-funcs is doing to stderr
 * If you have code you would not like scanned, you can use the `-ignore` flag to prune out matching paths
 * You can avoid calling the oracle during execution by passing in your own callgraph file in json format. See the oracle documentation for more info: https://docs.google.com/document/d/1SLk36YRjjMgKqe490mSRzOPYEDe0Y_WQNRv-EiFYUyw/view
 * Go oracle is not always correct, and may crash or improperly parse certain code
 * Go-unused-funcs only scans declarations. Functions created in the manner of 
   ```
   ExportedFiveFunc := func() int {return 5} 
   ```
   are not caught with this tool

As an example
```bash
$ bin/go-unused-funcs -v -all src/github.com/3rf/go-unused-funcs
```
```
Collecting func declarations from source files
Parsed 2 source files
Running callgraph analysis on following packages:
	github.com/3rf/go-unused-funcs
	github.com/3rf/go-unused-funcs/unused
Scanning callgraph for unused functions

Two in 'src/github.com/3rf/go-unused-funcs/main.go'
isNotStandardLibrary in 'src/github.com/3rf/go-unused-funcs/unused/unused_func_finder.go'
```

####Common Issues

1. If you are getting an error message like `Error getting results from oracle: import "/Users/kyle/folders/pkg/something": cannot import absolute path`, this is usually an indication that you are running the oracle on a folder that is not in your current GOPATH. Go-unused-funcs must be run on source files within your current GOPATH.
2. If the Go oracle is complaining about incorrect code, but your code is usable by the compiler, then congratulations: you've found a bug in the oracle tool. This happens much more often than I would like it to; one of the best things you can do is file a bug against it at https://code.google.com/p/go/issues/list. One possible workaround is to use the `-ignore` option to avoid the offending package.
