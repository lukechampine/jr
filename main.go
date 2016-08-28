package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/rpc/jsonrpc"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `USAGE:
  jr ADDRESS:PORT METHOD [PARAMETER...]

Options:
  -no-format	do not format JSON output (default: false)`)
}

func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
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
	var params json.RawMessage
	if len(args) > 2 {
		params = json.RawMessage(args[2])
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
	err = cli.Call(args[1], &params, &reply)
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
