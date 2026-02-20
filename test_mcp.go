package main

import (
"fmt"
"reflect"
"github.com/mark3labs/mcp-go/server"
)

func main() {
fmt.Println(reflect.TypeOf(server.NewMCPServer("a", "b")))
fmt.Println("NewSSEServer exists?", reflect.TypeOf(server.NewSSEServer))
}
