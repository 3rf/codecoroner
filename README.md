codecoroner [![Build Status](https://travis-ci.org/3rf/codecoroner.svg)](https://travis-ci.org/3rf/codecoroner)
===============


######Version 1.2 by Kyle Erf, MIT License 


Leaving dead code in a large codebase with multiple libraries is difficult to avoid.
Things get moved around; functions get refactored, leaving helpers on their own; people miscommunicate. 

The easiest ways to detect dead code is through static analysis. 
Unfortunately, Go's current static analysis tools (`oracle`, `callgraph`, etc) do not make aggregation of unused functions as easy as it should be.
This tool, codecoroner, uses the output of the Go `ast`, `ssa`, `callgraph/rta`, and `types` libraries to find unused functions/methods in your codebase.

The existing [unused code detectors](https://github.com/remyoudompheng/go-misc/tree/master/deadcode) are quite useful, but only work on a small scale.
Codecoroner was developed with large, multi-package, multi-main projects in mind.
The tool will detect unused functions and variables across packages, allowing you to see if your internal packages have exported code sitting unused.
At MongoDB, it helps us keep our repositories clean and even caught a couple bugs.
So far, codecoronoer has found dead code in every large public Go project I point it at.


###Quick Start

First, grab the `codecoroner` binary by either cloning this git repository and building main.go or by running
```bash
go get github.com/3rf/codecoroner
```
which should install a `codecoroner` binary in `$GOPATH/bin`

Codecoroner has two modes: `funcs` and `idents`, which detect dead code using callgraph and identifier analysis, respectively.
Each has their own set of benefits.

####Funcs

The `funcs` command builds a graph of function calls within your codebase and checks which functions are definied, but not reachable, from your `main` packages. 
This method takes advantage of the `callgraph/rta` implementation, which is geared toward applications like dead code analysis.

To run a `funcs` analysis, you can do
```bash
codecoroner funcs ./...
```
in the root of your project, similarly to how you would use `golint`.

Your results will look something like
```
unused/testdata/mockmain.go:15:1: oldHelper
unused/testdata/pkg1/random_num.go:31:1: toUint
unused/testdata/pkg1/random_num.go:36:1: GenUInt
unused/testdata/pkg1/random_num.go:42:1: GenSix
unused/testdata/pkg2/kittens.go:13:1: Val
unused/testdata/pkg2/kittens.go:25:1: GrayKittenLink
```

As a note: the `funcs` command only detects the usage of top-level functions and methods declared in the `func myFunc(a string){...}` form.
It does not track usage of anonymous functions or functions declared as package variables in the `var myFunc = func(a string){...}`; however, the `idents` command can catch the latter case.


####Idents

The `idents` command is a more simplistic and broad form of analysis.
It looks at every declared non-local identifier (package variables, functions, parameters, struct fields, methods) and checks to see that those identifiers are used outside of their declaration. 
Identifier analysis will catch things like unused constants, struct fields, and methods--all across packages.

To run an `idents` analysis, you can do
```bash
codecoroner idents ./...
```
in the root of your project.

Your results will look something like
```
github.com/3rf/codecoroner/unused/testdata/pkg1/random_num.go:10:7: Number
github.com/3rf/codecoroner/unused/testdata/pkg1/random_num.go:13:5: AnotherNumber
github.com/3rf/codecoroner/unused/testdata/pkg1/random_num.go:36:6: GenUInt
github.com/3rf/codecoroner/unused/testdata/pkg1/random_num.go:42:6: GenSix
github.com/3rf/codecoroner/unused/testdata/pkg2/kittens.go:11:25: field
github.com/3rf/codecoroner/unused/testdata/pkg2/kittens.go:13:7: ut
github.com/3rf/codecoroner/unused/testdata/pkg2/kittens.go:13:22: (unusedType).Val
github.com/3rf/codecoroner/unused/testdata/pkg2/kittens.go:25:6: GrayKittenLink
```

The `idents` command has more false positives and negatives than `funcs`. 
One reason for this is that `idents` does not build an execution graph, and so will not acknowledge code that is accessed through an interface, or catch unused code that is used cyclically but unreachable by main (e.g. `FuncA()` and `FuncB()` can call each other but nothing externally calls either of them).


### Full Usage

In addition to a command, the `codecoroner` executable requires a set of files as an argument.
You can pass in individual files and folders, or pass the current directory and its contents with `./...` in the style of the `go` program.
Codecoroner will automatically see what packages the files you give it belong to and include them in the dead code analysis. 
This API is designed to play nice with existing go tools and makes sense to me, but if you would prefer a different API, I'd be happy to hear you out.

Note that both modes will only report dead code for the packages/files you've passed to the tool.
Imports will be automatically detected so that callgraphs can be generated, but dead code inside those imports will not be reported.

#### Flags

#####-v
```
codecoroner -v funcs ./...
```

The verbose flag, `-v`, will print log messages to help you follow and troubleshoot your dead code analysis.
For example, running `codecoroner` in the root of its repository produces:
```
$ ./codecoroner -v funcs ./...
Collecting declarations from source files
Found pkg github.com/3rf/codecoroner
Ignoring path 'unused/funcs_test.go'
Ignoring path 'unused/idents_test.go'
Found pkg github.com/3rf/codecoroner/unused/testdata
Ignoring path 'unused/testdata/pkg1/random_num_test.go'
Parsed 8 source files
Running callgraph analysis on following packages:
	github.com/3rf/codecoroner
	github.com/3rf/codecoroner/unused/testdata
Running loader
Scanning callgraph for unused functions

unused/testdata/mockmain.go:15:1: oldHelper
unused/testdata/pkg1/random_num.go:31:1: toUint
unused/testdata/pkg1/random_num.go:36:1: GenUInt
unused/testdata/pkg1/random_num.go:42:1: GenSix
unused/testdata/pkg2/kittens.go:13:1: Val
unused/testdata/pkg2/kittens.go:25:1: GrayKittenLink
```

#####-tests
```
codecoroner -tests funcs ./...
```

The `-tests` flag includes test files and packages in your analysis. 
Doing this allows you to test main-less libraries and detect dead test helper code.


#####-ignore
```
codecoroner -ignore vendor,testdata funcs ./...
```

The `-ignore` flag accepts a comma-separated list of strings.
If any of the listed strings matches part of a filepath during scanning, that file will be ignored and excluded from the analysis.
This flag is a simple way to ignore vendored code without complicating the codecoroner's file argument.

#####-tags
```
codecornor -tags debug funcs ./...
```

The `-tags` flag lets you pass build tags like you would during a regular `go build`. 
If your codebase uses flags, note that unbuilt files may show up as dead code.

#### Troubleshooting

Some notes that may help with troubleshooting:
 * Make sure your code can actually compile with `go build` before running codecoroner on it.
 * If you have a vendoring system involving multiple GOPATHs, codecoroner should still work. In general, if you can execute `go build` from your current directory, you can run `codecoroner ./...` sucessfully.

When in doubt, file a GitHub issue and I'll be happy to help.


#### The Future
It would be great to drop the `idents` command and `funcs` commands altogether and do everything with SSA analysis.
None of the callgraph packages make the usage of non-function identifiers accessible, so it'll require haking at an existing implementation or building another callgraph package from scratch.
