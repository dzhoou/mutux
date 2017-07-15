package mutux

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// Mutux a mutable server that can be set at runtime to return any message at any URL.
type Mutux struct {
	Port      int
	Listener  *net.Listener
	Server    *http.Server
	Pathmsg   map[string]string
	Headers   map[string]string
	AllowPOST *bool
}

// Start start server in current process
func (m *Mutux) Start() error {
	return (*m.Server).Serve((*m.Listener).(*net.TCPListener))
}

// StartAsync start server in go routine
func (m *Mutux) StartAsync() {
	go (*m.Server).Serve((*m.Listener).(*net.TCPListener))
}

// AddPathMsg add message to a URL path
func (m *Mutux) AddPathMsg(path, msg string) {
	m.Pathmsg[path] = msg
}

// DelPathMsg delete msg from a URL path
func (m *Mutux) DelPathMsg(path string) {
	delete(m.Pathmsg, path)
}

// AddHeader add header to all GET responses
func (m *Mutux) AddHeader(name, value string) {
	m.Pathmsg[name] = value
}

// DelHeader delete header from all GET responses
func (m *Mutux) DelHeader(name string) {
	delete(m.Headers, name)
}

// EnablePOST enable modifying path message by POST
func (m *Mutux) EnablePOST() {
	*m.AllowPOST = true
}

// DisablePOST disable modifying path message by POST
func (m *Mutux) DisablePOST() {
	*m.AllowPOST = false
}

//NewMutux creates a new instance of Mutux server
func NewMutux(port int) (*Mutux, error) {
	pathmsg := map[string]string{}
	headers := map[string]string{
		"Content-type": "application/json",
	}
	allowPOST := true
	r := mux.NewRouter()
	// GET handler returns message for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]+}`, func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		vars := mux.Vars(r)
		name := vars["name"]
		msg, exists := pathmsg[name]
		if !exists {
			fmt.Fprintf(w, `{"error":"msg does not exist"}`)
			return
		}
		fmt.Fprintf(w, msg)
	}).Methods("GET")
	// POST handler stores message body for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]+}`, func(w http.ResponseWriter, r *http.Request) {
		if !allowPOST {
			return
		}
		vars := mux.Vars(r)
		name := vars["name"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
			return
		}
		pathmsg[name] = string(body)
		fmt.Fprintf(w, "success")
	}).Methods("POST")

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	mutux := Mutux{
		Port:      port,
		Server:    server,
		Listener:  &listener,
		Pathmsg:   pathmsg,
		Headers:   headers,
		AllowPOST: &allowPOST,
	}

	return &mutux, nil
}
