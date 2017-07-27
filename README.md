mutux-golang
============

![mutux Logo](http://imgur.com/7XnVIhM.png)

### `Mutux` creates a mutable message server, that can be modified at runtime via:

### 1. Program
```go
mutux.AddPathMsgAndStatus("hello", "Hello, world!", 200)
```
### 2. External PUT
```
PUT /hello
{"message":"Hello, world!", "status":200}
```

### `Mutux` also allows custom handler functions to be added on-the-fly to the server.
```go
fn := func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is coming from a custom handler function.")
}
mutuxServer.AddHandlerFunc(`/myfunc`, &fn, []string{"GET"})
```

### See also
 * [example/main.go](https://github.com/dzhoou/mutux/blob/master/example/main.go) -- example code
 * [mutux.go](https://github.com/dzhoou/mutux/blob/master/mutux.go) -- list of functions

### Dependency 
 * [gorilla/mux](https://github.com/gorilla/mux/)
