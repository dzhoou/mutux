mutux-golang
============

![mutux Logo](http://imgur.com/7XnVIhM.png)

#### `Mutux` creates a mutable message server, that can be modified at runtime via:

#### External POST
```
POST /hello
{"message":"Hello, world!", "status":200}
```
#### Program
```go
mutux.AddPathMsgAndStatus("hello", `{"message":"Hello, world!"}`, 200)
```

### See also
 * [example/main.go](https://github.com/dzhoou/mutux/blob/master/example/main.go) -- example code
 * [mutux.go](https://github.com/dzhoou/mutux/blob/master/mutux.go) -- list of functions

### Dependency 
  * [gorilla/mux](https://github.com/gorilla/mux/)
