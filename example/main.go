package main

import (
	"fmt"

	mutux "github.com/dzhoou/mutux"
)

func main() {
	fmt.Println("Starting go-mutux reflection server")
	mutux, err := mutux.NewMutux(8080)
	if err != nil {
		fmt.Println("Error initiating go-mutux reflection server: " + err.Error())
		return
	}
	mutux.StartAsync()
	mutux.AddPathMsg("hello", `{"message":"hello world!"}`)
	mutux.AddHeader("Content-Type", "application/json")
	// "select {}" hangs main program allowing server to run
	select {}
}
