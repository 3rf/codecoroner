go-unused-funcs
===============

Leaving dead code in a large codebase with multiple libraries is difficult to avoid.
Things get moved around; functions get refactored, leaving helpers on their own; people miscommunicate. 

One of the easiest ways to detect dead code is through static analysis. 
Unfortunately, Go's current static analysis tools (the oracle) do not make aggregation of unused functions as easy as it should be.
This tool, go-unused-funcs, uses the output of the Go oracle library to find unused functions/methods in your codebase. 

###Quick Start

First, grab the go-unused-funcs binary by either cloning this git repository and building main.go or by running
```bash
go get github.com/3rf/go-unused-funcs
```
which should install a go-unused-funcs binary in `$GOPATH/bin`

Go-unused-funcs takes in a set of files/folders, and scans them for functions and packages.
Every "main" package is collected and given to the go oracle for callgraph analysis.
The tool then checks all of the scanned function declarations against the callgraph to find unused ones.
To run the go-unused-funcs, simply call `go-unused-funcs` along with a set of files. For example:
```bash
go-unused-funcs src/github.com/3rf/go-unused-funcs/* 
```
will check the go-unused-funcs source folder for unused functions.

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

Notes:
 * If you would like to include test files in the callgraph (so that functions called only by tests are not marked as unused) run the tool with `-all`
 * Run with `-v` to print a log of what go-unused-funcs is doing to stderr
 * If you have code you would not like scanned, you can use the `-ignore` flag to prune out matching paths
 * You can avoid calling the oracle during execution by passing in your own callgraph file in json format. See the oracle documentation for more info: https://docs.google.com/document/d/1SLk36YRjjMgKqe490mSRzOPYEDe0Y_WQNRv-EiFYUyw/view
