package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

func usage() {
	os.Stderr.WriteString(`USAGE:
  jr ADDRESS:PORT METHOD [PARAMETER...]

Parameters are key/value pairs. They come in two forms:

  key=value     (for strings; result is "key":"value")
  key:=value    (for raw JSON; result is "key":value)

Options:
  -no-format	do not format JSON output (default: false)`)
}

func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

// parse arguments of the form bar, foo=bar, or foo:=3.
func parseArgs(args []string) json.RawMessage {
	// single, unkeyed argument
	if len(args) == 1 && !strings.ContainsRune(args[0], '=') {
		arg := args[0]
		if len(arg) > 0 && arg[0] == ':' {
			// raw JSON
			arg = arg[1:]
		} else {
			// unquoted string
			arg = strconv.Quote(arg)
		}
		return json.RawMessage(arg)
	}

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
	return json.RawMessage("{" + strings.Join(args, ",") + "}")
}

func main() {
	ugly := flag.Bool("no-format", false, "do not format JSON output")
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(2)
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
		params = json.RawMessage(stdin)
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
	if !json.Valid([]byte(reply)) {
		die("Call returned invalid JSON:", string(reply))
	}

	// print response, formatting if requested
	buf := new(bytes.Buffer)
	if *ugly {
		buf.Write([]byte(reply))
	} else {
		// reply is already validated, so no error possible
		json.Indent(buf, []byte(reply), "", "\t")
	}
	fmt.Println(buf)
}
