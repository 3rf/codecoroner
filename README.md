go-unused-funcs
===============

This tool uses the output of the go oracle to find unused
functions/methods in your codebase. It is hard to use,
but that is mostly because go oracle is hard to use. Woo.

###Instructions

First, you have to use the go oracle to generate a callgraph of
your packages in json format. For info on how to do this visit
the go oracle docs: https://docs.google.com/document/d/1SLk36YRjjMgKqe490mSRzOPYEDe0Y_WQNRv-EiFYUyw

The command will looks something like:
```
oracle -format=json callgraph path/for/pkgmain path/for/otherpkg ... > cg.json
```

Some notes:
  * You have to manually pass in all of the packages you wish to analyze. Sorry if you have a large codebase
  * Be sure to pipe the output into a file so go-unused-funcs can use it
  * Future versions of this tool will handle the oracle for you
  * You might think "oracle" is a bad name for a pkg that doesn't interface with the Oracle RDBMS, you are right to think that.

Once you've piped the json output to a file, it's time to run go-unused-funcs:
```
go-unused-funcs -calljson cg.json *
```

You may also substitute the star to be any file(s) you want to test.

The output should resemble:
```
getTaskFromRequest in apiserver/api.go NOT USED
fetchPatch in apiserver/patch_api.go NOT USED
ValidateIdToken in auth/oauth.go NOT USED
```

More notes:
  * Functions used in tests and nowhere else are counted in the callgraph, you can work around this by deleting all of your test files and then running the oracle and go-unused-funcs again. Hope you have version control!
  * This is not completely tested. I have yet to see any false positives, but they may be out there
  * One common pitfall is not including packages in the call to the oracle, double check your package list if you feel too many unused functions are being listed
