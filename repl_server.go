package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/dop251/goja"
)

//go:embed repl.js
var prelude string

func replServer(shell *Shell, logger *log.Logger) {
	vm := goja.New()
	globals := map[string]any{}
	shell.App.ExposeState(globals)
	vm.Set("shell", shell)
	vm.Set("g", globals)
	vm.RunString(prelude)
	listener, err := net.Listen("tcp", "127.0.0.1:8088")
	if err != nil {
		logger.Printf("failed to listen: %s\n", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Printf("failed to accept: %s\n", err)
			continue
		}
		logger.Printf("accepted connection from %s\n", conn.RemoteAddr())
		go repl(vm, conn)
	}
}

func printFunc(conn io.Writer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			fmt.Fprintf(conn, "%s\n", arg.String())
		}
		return goja.Undefined()
	}
}

func repl(vm *goja.Runtime, conn net.Conn) {
	lines := bufio.NewScanner(conn)
	conn.Write([]byte("> "))
	vm.Set("print", printFunc(conn))
	mode := "inspect"
	enc := json.NewEncoder(conn)
	for lines.Scan() {
		line := lines.Text()
		if line == "exit" {
			break
		}
		if strings.HasPrefix(line, ",mode ") {
			fields := strings.Split(line, " ")
			if len(fields) > 1 {
				mode = fields[1]
			}
			conn.Write([]byte("\n> "))
			continue
		}
		result, err := vm.RunString(line)
		if err != nil {
			fmt.Fprintf(conn, "error: %s\n", err)
			continue
		}
		if mode == "json" {
			enc.Encode(result)
		} else {
			conn.Write([]byte(result.ToString().String()))
		}
		conn.Write([]byte("\n> "))
	}
	conn.Close()
}
