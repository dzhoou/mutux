package main

import (
	"fmt"

	mutux "github.com/dzhoou/mutux"
)

func main() {
	fmt.Println("Starting Mutux server")
	mutuxServer, err := mutux.NewMutux(8080)
	if err != nil {
		fmt.Println("Error initiating Mutux server: " + err.Error())
		return
	}

	mutuxServer.StartAsync()
	defer mutuxServer.Stop()
	mutuxServer.AddPathMsg("hello", `{"message":"Hello, world!"}`)
	mutuxServer.AddHeader("Content-Type", "application/json")
	// "select {}" hangs main program allowing server to run
	select {}
}
