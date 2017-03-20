package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EndpointMethods define the optional methods that could
// be "overloaded" by any struct that defines an endpoint.
type EndpointMethods interface {
	GET(http.ResponseWriter, *http.Request)
	PUT(http.ResponseWriter, *http.Request)
	POST(http.ResponseWriter, *http.Request)
	DELETE(http.ResponseWriter, *http.Request)
	OPTIONS(http.ResponseWriter, *http.Request)
	DecorateJSON() bool
	DecorateLOG() bool
}

// EndpointHandler is a default implementation of the EndpointMethods interface.
// It adds 2 methods to control how the HTTP methods get decorated.
// All methods for the same endpoint get the same decorations.
type EndpointHandler struct {
	AddLogDecorator  bool
	AddJSONDecorator bool
}

// GET default HTTP GET responder, HTTP 405 StatusMethodNotAllowed
func (e *EndpointHandler) GET(w http.ResponseWriter, req *http.Request) {
	httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
}

// PUT default HTTP PUT responder, HTTP 405 StatusMethodNotAllowed
func (e *EndpointHandler) PUT(w http.ResponseWriter, req *http.Request) {
	httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
}

// POST default HTTP POST responder, HTTP 405 StatusMethodNotAllowed
func (e *EndpointHandler) POST(w http.ResponseWriter, req *http.Request) {
	httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
}

// DELETE default HTTP DELETE responder, HTTP 405 StatusMethodNotAllowed
func (e *EndpointHandler) DELETE(w http.ResponseWriter, req *http.Request) {
	httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
}

// OPTIONS default HTTP OPTIONS responder, HTTP 405 StatusMethodNotAllowed
func (e *EndpointHandler) OPTIONS(w http.ResponseWriter, req *http.Request) {
	httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
}

// DecorateJSON returns if the methods in this interface should have a JSON formatted response or not, default NO
func (e *EndpointHandler) DecorateJSON() bool { return false }

// DecorateLOG returns if the methods in this interface should have a Print log or not, default NO
func (e *EndpointHandler) DecorateLOG() bool { return false }

// Index implementation of the default EndpointHandler, it can "override" the default methods
// to enable functionality for its endpoints
type Index struct {
	EndpointHandler
}

// GET is now allowed on this endpoint
func (e *Index) GET(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "{\"msg\":\"GET!\"}\n")
}

// DecorateJSON returns if the methods in this interface should have a JSON formatted response or not depending on the set bool
func (idx *Index) DecorateJSON() bool { return idx.AddJSONDecorator }

// DecorateLOG returns if the methods in this interface should have a Print log or not depending on the set bool
func (idx *Index) DecorateLOG() bool { return idx.AddLogDecorator }

// ErrStruct basic json formatted struct
type ErrStruct struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// httpError is an convenience function
func httpError(code int, errString string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(ErrStruct{Code: code, Message: errString})
	if err != nil {
		io.WriteString(w, fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", code, errString))
	}
}

// urlHandler lookups a http.HandlerFunc by HTTP Method
type urlHandlerMap map[string]http.HandlerFunc

// methodUrlHandler lookups a http.HandlerFunc by HTTP Method
type methodURLHandlerMap map[string]urlHandlerMap

// MyMux is our own special muxer
// All it does is keep track of which endpoints and HTTP methods belong together.
// Nothig fancy with path matching.
type MyMux struct {
	handlers methodURLHandlerMap
}

// ServeHTTP basic HTTP Handler
// Register the handler for the proper path and method
func (m MyMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Match Method to stored handlers
	if urlH := m.handlers[req.URL.Path]; urlH != nil {
		if h := urlH[req.Method]; h != nil {
			h(w, req)
		} else {
			httpError(http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed on %s", req.Method, req.URL.Path), w)
		}
	} else {
		httpError(http.StatusNotFound, fmt.Sprint("missing"), w)
	}
}

// RegisterHandler registers new handlers
func (m *MyMux) RegisterHandler(method string, path string, h http.HandlerFunc) {
	if m.handlers == nil {
		m.handlers = make(methodURLHandlerMap, 0)
	}

	if urlH := m.handlers[path]; urlH == nil {
		m.handlers[path] = make(urlHandlerMap, 1)
	}
	m.handlers[path][method] = h

}

// Register takes an instance of a struct (or such) which implements the EndpointMethods interface.
// Decorate the handlers properly.
func (m *MyMux) Register(path string, edph EndpointMethods) {
	if m.handlers == nil {
		m.handlers = make(methodURLHandlerMap, 0)
	}

	if urlH := m.handlers[path]; urlH == nil {
		m.handlers[path] = make(urlHandlerMap, 1)
	}

	m.handlers[path][http.MethodGet] = edph.GET
	m.handlers[path][http.MethodPut] = edph.PUT
	m.handlers[path][http.MethodPost] = edph.POST
	m.handlers[path][http.MethodDelete] = edph.DELETE
	m.handlers[path][http.MethodOptions] = edph.OPTIONS

	if edph.DecorateJSON() {
		m.handlers[path][http.MethodGet] = jsonDecorator(m.handlers[path][http.MethodGet])
		m.handlers[path][http.MethodPut] = jsonDecorator(m.handlers[path][http.MethodPut])
		m.handlers[path][http.MethodPost] = jsonDecorator(m.handlers[path][http.MethodPost])
		m.handlers[path][http.MethodDelete] = jsonDecorator(m.handlers[path][http.MethodDelete])
		m.handlers[path][http.MethodOptions] = jsonDecorator(m.handlers[path][http.MethodOptions])
	}

	if edph.DecorateLOG() {
		m.handlers[path][http.MethodGet] = printDecorator(m.handlers[path][http.MethodGet])
		m.handlers[path][http.MethodPut] = printDecorator(m.handlers[path][http.MethodPut])
		m.handlers[path][http.MethodPost] = printDecorator(m.handlers[path][http.MethodPost])
		m.handlers[path][http.MethodDelete] = printDecorator(m.handlers[path][http.MethodDelete])
		m.handlers[path][http.MethodOptions] = printDecorator(m.handlers[path][http.MethodOptions])
	}

}

// GET alias for RegisterHandler w/ GET argument
func (m *MyMux) GET(path string, h http.HandlerFunc) {
	m.RegisterHandler(http.MethodGet, path, h)
}

// PUT alias for RegisterHandler w/ PUT argument
func (m *MyMux) PUT(path string, h http.HandlerFunc) {
	m.RegisterHandler(http.MethodPut, path, h)
}

// POST alias for RegisterHandler w/ POST argument
func (m *MyMux) POST(path string, h http.HandlerFunc) {
	m.RegisterHandler(http.MethodPost, path, h)
}

// printDecorator just prints the request to stdout
func printDecorator(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("%s \"%s\" -> %s\n", req.Method, req.URL.Path, req.UserAgent())
		h(w, req)
	}
}

// jsonDecorator just adds application/json to the response header
func jsonDecorator(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		h(w, req)
	}
}

func main() {
	fmt.Println("Starting server")
	idx := Index{EndpointHandler{true, true}}
	mux := MyMux{}

	// Register my endpoint struct
	mux.Register("/", &idx)

	// We can also decorate and add the http.HandlerFuncs manually
	mux.GET("/get", printDecorator(jsonDecorator(idx.GET)))
	mux.POST("/post", printDecorator(jsonDecorator(idx.POST)))
	mux.GET("/another/get", printDecorator(jsonDecorator(idx.GET)))

	// Start server
	err := http.ListenAndServe(":900", mux)
	fmt.Printf("Closing with err: %v\n", err)

}
