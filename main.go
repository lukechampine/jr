package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/rpc/jsonrpc"
	"os"
)

var haveStdin bool = func() bool {
	// detect whether the user is piping data to stdin
	// (not perfect; doesn't work with e.g. /dev/zero)
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}()

func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		die(`USAGE:
  jr ADDRESS:PORT METHOD [PARAMETER...]`)
	}

	// parse params, if supplied
	var params json.RawMessage
	if len(os.Args) > 3 {
		params = json.RawMessage(os.Args[3])
	} else if haveStdin {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			die("Couldn't read from stdin:", err)
		}
		params = json.RawMessage(stdin)
	}

	// connect to server
	cli, err := jsonrpc.Dial("tcp", os.Args[1])
	if err != nil {
		die("Couldn't connect to server:", err)
	}
	defer cli.Close()

	// call
	var reply json.RawMessage
	err = cli.Call(os.Args[2], &params, &reply)
	if err != nil {
		die("Call failed:", err)
	}

	// print
	fmt.Println(string(reply))
}
