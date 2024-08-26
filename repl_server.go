package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"

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

func repl(vm *goja.Runtime, conn net.Conn) {
	lines := bufio.NewScanner(conn)
	conn.Write([]byte("> "))
	enc := json.NewEncoder(conn)
	for lines.Scan() {
		line := lines.Text()
		if line == "exit" {
			break
		}
		result, err := vm.RunString(line)
		if err != nil {
			fmt.Fprintf(conn, "error: %s\n", err)
			continue
		}
		enc.Encode(result)
		conn.Write([]byte("\n> "))
	}
	conn.Close()
}
