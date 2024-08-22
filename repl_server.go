package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"

	"github.com/robertkrimen/otto"
)

//go:embed repl.js
var prelude string

func replServer(app *App) {
	vm := otto.New()
	globals := map[string]any{}
	app.ExposeState(globals)
	vm.Set("g", globals)
	vm.Run(prelude)
	listener, err := net.Listen("tcp", "127.0.0.1:8088")
	if err != nil {
		fmt.Printf("[repl] failed to listen: %s\n", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[repl] failed to accept: %s\n", err)
			continue
		}
		fmt.Printf("[repl] accepted connection from %s\n", conn.RemoteAddr())
		go repl(vm, conn)
	}
}

func repl(vm *otto.Otto, conn net.Conn) {
	lines := bufio.NewScanner(conn)
	conn.Write([]byte("> "))
	enc := json.NewEncoder(conn)
	for lines.Scan() {
		line := lines.Text()
		if line == "exit" {
			break
		}
		result, err := vm.Run(line)
		if err != nil {
			fmt.Fprintf(conn, "error: %s\n", err)
			continue
		}
		enc.Encode(result)
		conn.Write([]byte("\n> "))
	}
	conn.Close()
}
