package mutux

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"github.com/gorilla/mux"
)

// Mutux a mutable server that can be set at runtime to return any message at any URL.
type Mutux struct {
	Address  string
	Listener *net.Listener
	Server   *http.Server
	Pathmsg  map[string]Message
	Headers  map[string]string
	AllowPUT *bool
}

// Message store message and status to return for a given path
type Message struct {
	Msg    *string `json:"message"`
	Status *int    `json:"status"`
}

func (m *Mutux) remakeListener() error {
	if m == nil {
		return nil
	}
	listener, err := net.Listen("tcp", m.Address)
	if err != nil {
		return err
	}
	m.Listener = &listener
	return nil
}

// Start start Mutux server in current process
func (m *Mutux) Start() error {
	if m == nil {
		return nil
	}
	if m.Listener == nil {
		err := m.remakeListener()
		if err != nil {
			return err
		}
	}
	fmt.Println("starting server")
	return (*m.Server).Serve((*m.Listener).(*net.TCPListener))
}

// StartAsync start Mutux server in go routine
func (m *Mutux) StartAsync() error {
	if m == nil {
		return nil
	}
	if m.Listener == nil {
		err := m.remakeListener()
		if err != nil {
			return err
		}
	}
	fmt.Println("starting server async")
	go (*m.Server).Serve((*m.Listener).(*net.TCPListener))
	return nil
}

// Stop close Mutux server
func (m *Mutux) Stop() error {
	if m == nil {
		return nil
	}
	if m.Server != nil {
		fmt.Println("closing server")
		m.Server.Shutdown(nil)
		m.Listener = nil
	}
	return nil
}

// AddPathMsg add message to a URL path
func (m *Mutux) AddPathMsg(path, msg string) {
	if m == nil {
		return
	}
	status := 200
	// Need to strip preceding "/", as well as any URL parameters.
	// Support for URL params will be added later.
	i := 0
	pathlen := len(path)
	for i < pathlen && path[i] == '/' {
		i++
	}
	if i > 0 {
		path = path[i:pathlen]
	}
	path = strings.Split(path, "?")[0]
	fmt.Println(fmt.Sprintf("adding path /%s", path))
	m.Pathmsg[path] = Message{
		Msg:    &msg,
		Status: &status,
	}
}

// AddPathMsgAndStatus add message to a URL path, with specified status code
func (m *Mutux) AddPathMsgAndStatus(path, msg string, status int) {
	if m == nil {
		return
	}
	// Need to strip preceding "/", as well as any URL parameters.
	// Support for URL params will be added later.
	i := 0
	pathlen := len(path)
	for i < pathlen && path[i] == '/' {
		i++
	}
	if i > 0 {
		path = path[i:pathlen]
	}
	path = strings.Split(path, "?")[0]
	fmt.Println(fmt.Sprintf("adding path /%s with status %d", path, status))
	m.Pathmsg[path] = Message{
		Msg:    &msg,
		Status: &status,
	}
}

// DelPathMsg delete msg from a URL path
func (m *Mutux) DelPathMsg(path string) {
	if m == nil {
		return
	}
	delete(m.Pathmsg, path)
}

// AddHeader add header to all GET and POST responses
func (m *Mutux) AddHeader(name, value string) {
	if m == nil {
		return
	}
	m.Headers[name] = value
}

// DelHeader delete header from all GET and POST responses
func (m *Mutux) DelHeader(name string) {
	if m == nil {
		return
	}
	delete(m.Headers, name)
}

// EnablePUT enable modifying path message by PUT
func (m *Mutux) EnablePUT() {
	if m == nil {
		return
	}
	*m.AllowPUT = true
}

// DisablePUT disable modifying path message by PUT
func (m *Mutux) DisablePUT() {
	if m == nil {
		return
	}
	*m.AllowPUT = false
}

//NewMutux creates a new instance of Mutux server with port number specified
func NewMutux(port int) (*Mutux, error) {
	return NewMutuxWithAddr(fmt.Sprintf(":%d", port))
}

//NewMutuxWithAddr creates a new instance of Mutux server with string address specified
func NewMutuxWithAddr(addr string) (*Mutux, error) {
	pathmsg := map[string]Message{}
	headers := map[string]string{
		"Content-type": "application/json",
	}
	allowPUT := true
	r := mux.NewRouter()

	// GET handler returns message for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]*}`, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		msg, exists := pathmsg[name]
		if !exists {
			http.Error(w, "404 page not found", 404)
			return
		}
		w.WriteHeader(*msg.Status)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		fmt.Fprintf(w, *msg.Msg)
	}).Methods("GET")

	// POST handler returns message for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]*}`, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		msg, exists := pathmsg[name]
		if !exists {
			http.Error(w, "404 page not found", 404)
			return
		}
		w.WriteHeader(*msg.Status)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		fmt.Fprintf(w, *msg.Msg)
	}).Methods("POST")

	// PUT handler stores message body for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]*}`, func(w http.ResponseWriter, r *http.Request) {
		if !allowPUT {
			return
		}
		vars := mux.Vars(r)
		name := vars["name"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading body: %s", err.Error()), 500)
			return
		}
		putmsg := Message{}
		err = json.Unmarshal(body, &putmsg)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling body: %s", err.Error()), 500)
			return
		}
		if putmsg.Msg == nil {
			http.Error(w, fmt.Sprintf("Error: message is empty"), 500)
			return
		}
		if putmsg.Status == nil {
			status := 200
			putmsg.Status = &status
		}
		fmt.Println("adding path: /" + name)
		pathmsg[name] = putmsg
		fmt.Fprintf(w, "success")
	}).Methods("PUT")

	// OPTIONS handler handles browser CORS preflight
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]*}`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT")
	}).Methods("OPTIONS")

	server := &http.Server{Addr: addr, Handler: r}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	mutux := Mutux{
		Address:  addr,
		Server:   server,
		Listener: &listener,
		Pathmsg:  pathmsg,
		Headers:  headers,
		AllowPUT: &allowPUT,
	}

	return &mutux, nil
}
