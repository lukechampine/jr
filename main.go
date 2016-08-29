package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/rpc/jsonrpc"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, `USAGE:
  jr ADDRESS:PORT METHOD [PARAMETER...]

Parameters are key/value pairs. They come in two forms:

  key=value     (for strings; result is "key":"value")
  key:=value    (for raw JSON; result is "key":value)

Options:
  -no-format	do not format JSON output (default: false)`)
	os.Exit(1)
}

func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

// parse arguments of the form foo=bar or foo:=3.
// NOTE: the arguments are all packed into one object, which is sent as the
// only parameter. Not sure if this is standard or just a quirk of the Go
// jsonrpc package.
func parseArgs(args []string) json.Marshaler {
	for i, arg := range args {
		eq := strings.IndexByte(arg, '=')
		if eq == -1 {
			die("Invalid argument:", arg)
		}
		if arg[eq-1] == ':' {
			// raw JSON; only quote the key
			args[i] = fmt.Sprintf("%q:%s", arg[:eq-1], arg[eq+1:])
		} else {
			// unquoted string; add quotes
			args[i] = fmt.Sprintf("%q:%q", arg[:eq], arg[eq+1:])
		}
	}
	js := json.RawMessage("{" + strings.Join(args, ",") + "}")
	return &js
}

func main() {
	ugly := flag.Bool("no-format", false, "do not format JSON output")
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		usage()
	}

	// detect whether the user is piping data to stdin
	// (not perfect; doesn't work with e.g. /dev/zero)
	stat, _ := os.Stdin.Stat()
	haveStdin := (stat.Mode() & os.ModeCharDevice) == 0

	// parse params, if supplied
	var params json.Marshaler
	if len(args) > 2 {
		params = parseArgs(args[2:])
	} else if haveStdin {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			die("Couldn't read from stdin:", err)
		}
		js := json.RawMessage(stdin)
		params = &js // RawMessage needs pointer receiver
	}

	// connect to server
	cli, err := jsonrpc.Dial("tcp", args[0])
	if err != nil {
		die("Couldn't connect to server:", err)
	}
	defer cli.Close()

	// call
	var reply json.RawMessage
	err = cli.Call(args[1], params, &reply)
	if err != nil {
		die("Call failed:", err)
	}

	// print response, formatting if requested
	if *ugly {
		fmt.Println(string(reply))
	} else {
		buf := new(bytes.Buffer)
		err = json.Indent(buf, []byte(reply), "", "\t")
		if err != nil {
			die("Couldn't format response:", err)
		}
		fmt.Println(buf)
	}
}
