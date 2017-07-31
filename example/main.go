package main

import (
	"fmt"

	"net/http"

	mutux "github.com/dzhoou/mutux"
)

func main() {

	mutuxServer, err := mutux.NewMutux(6666)
	if err != nil {
		fmt.Println("Error initiating Mutux server: " + err.Error())
		return
	}

	fmt.Println("Starting Mutux server")
	mutuxServer.Start()

	// Adds a message to path /hello
	mutuxServer.AddPathMsg("hello", `{"message":"Hello, world!"}`)

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "myfunc")
	}
	// Adds function to path /myfunc
	mutuxServer.AddHandlerFunc(`/myfunc`, &fn, []string{"GET"})
	// In order for added functions to take effect, the server has to be restarted
	// Alternatively, you can call mutuxServer.AddHandlerFuncAndRestart() to add a function and restart.
	mutuxServer.Restart()

	// "select {}" hangs main program allowing server to run
	select {}
}
