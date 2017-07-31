package mutux

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"net/http"

	"github.com/dzhoou/mutux/server19"
	"github.com/gorilla/mux"
)

// Mutux a mutable server that can be set at runtime to return any message at any URL.
type Mutux struct {
	Address            string
	Certfile           string
	Keyfile            string
	Listener           *net.Listener
	Server             *server19.Server
	Pathmsg            map[string]Message
	Headers            map[string]string
	AllowPUT           *bool
	Handler            *mux.Router
	CustomHandlerfuncs []Handlerfunc
	handlerfuncs       []Handlerfunc
}

// Message store message and status to return for a given path
type Message struct {
	Msg    *string `json:"message"`
	Status *int    `json:"status"`
}

// Handlerfunc store instances of handler function
type Handlerfunc struct {
	Route    string
	Methods  []string
	Function *func(w http.ResponseWriter, r *http.Request)
}

func (m *Mutux) remakeListener() error {
	if m == nil {
		return nil
	}
	fmt.Printf("remaking listener...")
	listener, err := net.Listen("tcp", m.Address)
	if err != nil {
		for i := 0; i < 100; i++ {
			time.Sleep(20 * time.Millisecond)
			listener, err = net.Listen("tcp", m.Address)
			if err == nil {
				break
			}
			fmt.Printf(".")
		}
	}
	if err != nil {
		fmt.Println("\nFailed to remake listener after retries")
		return err
	}
	fmt.Println("\nSuccess remaking listener.")
	m.Listener = &listener
	return nil
}

// StartAndHold start Mutux server in current process
func (m *Mutux) StartAndHold() error {
	if m == nil {
		return nil
	}
	if m.Listener == nil {
		err := m.remakeListener()
		if err != nil {
			return err
		}
	}
	fmt.Println("starting server in current process")
	if m.Certfile != "" && m.Keyfile != "" {
		return (*m.Server).ServeTLS((*m.Listener).(*net.TCPListener), m.Certfile, m.Keyfile)
	}
	return (*m.Server).Serve((*m.Listener).(*net.TCPListener))
}

// Start start Mutux server in go routine
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
	if m.Certfile != "" && m.Keyfile != "" {
		go (*m.Server).ServeTLS((*m.Listener).(*net.TCPListener), m.Certfile, m.Keyfile)
	} else {
		go (*m.Server).Serve((*m.Listener).(*net.TCPListener))
	}
	return nil
}

// Stop close Mutux server
func (m *Mutux) Stop() error {
	if m == nil {
		return nil
	}
	if m.Server != nil {
		fmt.Println("\nClosing server")
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

// AddHandlerFunc add user-defined handler func to path
func (m *Mutux) AddHandlerFunc(route string, f *func(w http.ResponseWriter, r *http.Request), methods []string) {
	if m == nil {
		return
	}
	// Add to func array
	m.CustomHandlerfuncs = append(m.CustomHandlerfuncs, Handlerfunc{
		Route:    route,
		Function: f,
		Methods:  methods,
	})
}

// AddHandlerFuncAndRestart add user-defined handler func to path, and restart the server for change to take effect
func (m *Mutux) AddHandlerFuncAndRestart(route string, f *func(w http.ResponseWriter, r *http.Request), methods []string) error {
	if m == nil {
		return nil
	}
	// Add to func array
	m.CustomHandlerfuncs = append(m.CustomHandlerfuncs, Handlerfunc{
		Route:    route,
		Function: f,
		Methods:  methods,
	})
	err := m.Restart()
	if err != nil {
		return fmt.Errorf("Failed to add handler function: %s", err.Error())
	}
	return nil
}

// ClearHandlerFunc delete all user-defined handler funcs, and restart the server for change to take effect
func (m *Mutux) ClearHandlerFunc() {
	if m == nil {
		return
	}
	// delete func array
	m.CustomHandlerfuncs = nil
}

// ClearHandlerFuncAndRestart delete all user-defined handler funcs, and restart the server for change to take effect
func (m *Mutux) ClearHandlerFuncAndRestart() error {
	if m == nil {
		return nil
	}
	// delete func array
	m.CustomHandlerfuncs = nil
	err := m.Restart()
	if err != nil {
		return fmt.Errorf("Failed to clear handler functions: %s", err.Error())
	}
	return nil
}

// Restart restart server (to load newly added functions to server host, for example)
func (m *Mutux) Restart() error {
	if m == nil {
		return nil
	}
	// recreate router
	r := mux.NewRouter()

	m.addHandlersToRouter(r)

	err := m.Stop()
	if err != nil {
		return fmt.Errorf("Failed to remake router: %s", err.Error())
	}
	m.Server = &server19.Server{}
	m.Server.Addr = m.Address
	m.Server.Handler = r
	err = m.Start()
	if err != nil {
		return fmt.Errorf("Failed to remake router: %s", err.Error())
	}
	return nil
}

func (m *Mutux) addHandlersToRouter(r *mux.Router) {
	// add custom funcs to router; they are added before the original funcs because otherwise the original funcs would override the custom funcs
	// add custom funcs in reverse order so that newly added funcs have higher processing priority
	for i := len(m.CustomHandlerfuncs) - 1; i >= 0; i-- {
		h := m.CustomHandlerfuncs[i]
		if h.Methods != nil {
			for _, m := range h.Methods {
				r.HandleFunc(h.Route, *h.Function).Methods(m)
			}
		} else {
			r.HandleFunc(h.Route, *h.Function)
		}
	}
	// add back original message funcs to router
	for _, h := range m.handlerfuncs {
		if h.Methods != nil {
			for _, m := range h.Methods {
				r.HandleFunc(h.Route, *h.Function).Methods(m)
			}
		} else {
			r.HandleFunc(h.Route, *h.Function)
		}
	}
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

	GETmessagefunc := func(w http.ResponseWriter, r *http.Request) {
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
		fmt.Printf("\nIn GET message handler, returning: \n%s\n", *msg.Msg)
		fmt.Fprintf(w, *msg.Msg)
	}
	POSTmessagefunc := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		msg, exists := pathmsg[name]
		if !exists {
			http.Error(w, `{"error":"404 page not found"}`, 404)
			return
		}
		w.WriteHeader(*msg.Status)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		fmt.Printf("\nIn POST message handler, returning: \n%s\n", *msg.Msg)
		fmt.Fprintf(w, *msg.Msg)
	}
	PUTmessagefunc := func(w http.ResponseWriter, r *http.Request) {
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
	}
	CORSfunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT")
	}

	handlerfuncs := []Handlerfunc{
		// GET handler returns message for any URL path
		Handlerfunc{
			Route:    `/{name:[a-zA-Z0-9=\-\/]*}`,
			Function: &GETmessagefunc,
			Methods:  []string{"GET"},
		},
		// POST handler returns message for any URL path
		Handlerfunc{
			Route:    `/{name:[a-zA-Z0-9=\-\/]*}`,
			Function: &POSTmessagefunc,
			Methods:  []string{"POST"},
		},
		// PUT handler updates message for any URL path
		Handlerfunc{
			Route:    `/{name:[a-zA-Z0-9=\-\/]*}`,
			Function: &PUTmessagefunc,
			Methods:  []string{"PUT"},
		},
		// OPTIONS handler handles browser CORS preflight
		Handlerfunc{
			Route:    `/{name:[a-zA-Z0-9=\-\/]*}`,
			Function: &CORSfunc,
			Methods:  []string{"OPTIONS"},
		},
	}

	server := &server19.Server{}
	server.Addr = addr
	server.Handler = r
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	mutux := Mutux{
		Address:      addr,
		Server:       server,
		Listener:     &listener,
		Pathmsg:      pathmsg,
		Headers:      headers,
		AllowPUT:     &allowPUT,
		Handler:      r,
		handlerfuncs: handlerfuncs,
	}

	mutux.addHandlersToRouter(r)

	return &mutux, nil
}
