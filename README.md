jr
==

`jr` is a JSON-RPC command-line client written in Go. It is based on
[jsonrpcake](https://github.com/joehillen/jsonrpcake), which was based on
[HTTPie](https://github.com/jkbrzt/httpie). Usage of `jr` should be familiar
to users of those programs.

`jr` does not support the `@` syntax or colored output of the aforementioned
programs. I'll add them if someone asks me to.

Installation
------------

```
go get github.com/lukechampine/jr
```

Usage
-----

```bash
# no hostname means localhost
$ jr :3000 hello
Hello, World!

# string parameter
$ jr :3000 hello name=Luke
Hello, Luke!

# bool parameter
$ jr :3000 hello name=Luke excited:=false
Hey, Luke.

# stdin is not processed; must be valid JSON
$ cat | jr :3000 hello
{
	"name": "Luke",
	"excited": false
}
^D
Hey, Luke.
```
