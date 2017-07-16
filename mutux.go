package mutux

import (
	"encoding/json"
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
	Pathmsg   map[string]Message
	Headers   map[string]string
	AllowPOST *bool
}

// Message store message and status to return for a given path
type Message struct {
	Msg    *string `json:"message"`
	Status *int    `json:"status"`
}

func (m *Mutux) remakeListener() error {
	addr := fmt.Sprintf(":%d", m.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.Listener = &listener
	return nil
}

// Start start Mutux server in current process
func (m *Mutux) Start() error {
	if m.Listener == nil {
		err := m.remakeListener()
		if err != nil {
			return err
		}
		return nil
	}
	return (*m.Server).Serve((*m.Listener).(*net.TCPListener))
}

// StartAsync start Mutux server in go routine
func (m *Mutux) StartAsync() error {
	if m.Listener == nil {
		err := m.remakeListener()
		if err != nil {
			return err
		}
	}
	go (*m.Server).Serve((*m.Listener).(*net.TCPListener))
	return nil
}

// Stop close Mutux server listener
func (m *Mutux) Stop() error {
	if m.Listener != nil {
		err := (*m.Listener).Close()
		if err != nil {
			return err
		}
		m.Listener = nil
	}
	return nil
}

// AddPathMsg add message to a URL path
func (m *Mutux) AddPathMsg(path, msg string) {
	status := 200
	m.Pathmsg[path] = Message{
		Msg:    &msg,
		Status: &status,
	}
}

// AddPathMsgAndStatus add message to a URL path, with specified status code
func (m *Mutux) AddPathMsgAndStatus(path, msg string, status int) {
	m.Pathmsg[path] = Message{
		Msg:    &msg,
		Status: &status,
	}
}

// DelPathMsg delete msg from a URL path
func (m *Mutux) DelPathMsg(path string) {
	delete(m.Pathmsg, path)
}

// AddHeader add header to all GET responses
func (m *Mutux) AddHeader(name, value string) {
	m.Headers[name] = value
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
	pathmsg := map[string]Message{}
	headers := map[string]string{
		"Content-type": "application/json",
	}
	allowPOST := true
	r := mux.NewRouter()

	// GET handler returns message for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]+}`, func(w http.ResponseWriter, r *http.Request) {
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

	// POST handler stores message body for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]+}`, func(w http.ResponseWriter, r *http.Request) {
		if !allowPOST {
			return
		}
		vars := mux.Vars(r)
		name := vars["name"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading body: %s", err.Error()), 500)
			return
		}
		postmsg := Message{}
		err = json.Unmarshal(body, &postmsg)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling body: %s", err.Error()), 500)
			return
		}
		if postmsg.Msg == nil {
			http.Error(w, fmt.Sprintf("Error: message is empty"), 500)
			return
		}
		if postmsg.Status == nil {
			status := 200
			postmsg.Status = &status
		}
		pathmsg[name] = postmsg
		fmt.Fprintf(w, "success")
	}).Methods("POST")

	// POST handler stores message body for any URL path
	r.HandleFunc(`/{name:[a-zA-Z0-9=\-\/]+}`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}).Methods("OPTIONS")

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
