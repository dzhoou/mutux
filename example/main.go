package main

import (
	"fmt"
	"net/http"

	mutux "github.com/dzhoou/mutux"
)

func main() {

	mutuxServer, err := mutux.NewMutux(8080)
	if err != nil {
		fmt.Println("Error initiating Mutux server: " + err.Error())
		return
	}

	fmt.Println("Starting Mutux server")
	mutuxServer.StartAsync()

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "myfunc")
	}
	mutuxServer.AddHandlerFunc(`/myfunc`, &fn, []string{"GET"})
	mutuxServer.AddPathMsg("hello", `{"message":"Hello, world!"}`)

	// "select {}" hangs main program allowing server to run
	select {}
}
