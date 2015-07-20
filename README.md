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

Codecoroner has two modes: `funcs` and `idents`, which detect dead code using callgraph and identifier analysis, respectively.

Examples:
```
codecoroner -tests -v funcs .

codecoroner idents .

codecoroner -ignore testdata,vendor funcs .
```
