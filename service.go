package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type urlHandler map[string]http.HandlerFunc
type methodURLHandler map[string]urlHandler

type EndpointMethods interface {
	GET(http.ResponseWriter, *http.Request)
	PUT(http.ResponseWriter, *http.Request)
	POST(http.ResponseWriter, *http.Request)
	DELETE(http.ResponseWriter, *http.Request)
	OPTIONS(http.ResponseWriter, *http.Request)
}
type EndpointHandler struct {
	AddLogDecorator  bool
	AddJSONDecorator bool
}

func (e *EndpointHandler) GET(_ http.ResponseWriter, _ *http.Request)     {}
func (e *EndpointHandler) PUT(_ http.ResponseWriter, _ *http.Request)     {}
func (e *EndpointHandler) POST(_ http.ResponseWriter, _ *http.Request)    {}
func (e *EndpointHandler) DELETE(_ http.ResponseWriter, _ *http.Request)  {}
func (e *EndpointHandler) OPTIONS(_ http.ResponseWriter, _ *http.Request) {}

type Index struct {
	EndpointHandler
}

func (e *Index) GET(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "{\"msg\":\"GET!\"}\n")
}

// MyMux is our own special muxer
type MyMux struct {
	handlers methodURLHandler
}

// ErrStruct basic json formatted struct
type ErrStruct struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func httpError(code int, errString string, w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(ErrStruct{Code: code, Message: errString})
	if err != nil {
		io.WriteString(w, fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", code, errString))
	}
}

// ServeHTTP basic HTTP Handler
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
		m.handlers = make(methodURLHandler, 0)
	}

	if urlH := m.handlers[path]; urlH == nil {
		m.handlers[path] = make(urlHandler, 1)
	}
	m.handlers[path][method] = h

}

// GET alias for RegisterHandler w/ GET argument
func (m *MyMux) GET(path string, h http.HandlerFunc) {
	m.RegisterHandler("GET", path, h)
}

// PUT alias for RegisterHandler w/ GET argument
func (m *MyMux) PUT(path string, h http.HandlerFunc) {
	m.RegisterHandler("PUT", path, h)
}

// POST alias for RegisterHandler w/ GET argument
func (m *MyMux) POST(path string, h http.HandlerFunc) {
	m.RegisterHandler("POST", path, h)
}

func printDecorator(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("%s \"%s\" %s  -> %s\n", req.Method, req.URL.RawQuery, req.URL.Fragment, req.UserAgent())
		h(w, req)
	}
}

func jsonDecorator(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		h(w, req)
	}
}

// index
func indexGET(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\"msg\":\"GET!\"}\n")
}

// index
func indexPOST(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\"msg\":\"POST!\"}\n")
}

func main() {
	fmt.Println("Starting server")
	idx := Index{EndpointHandler{true, true}}
	mux := MyMux{}
	mux.GET("/", printDecorator(jsonDecorator(idx.GET)))
	mux.POST("/", printDecorator(jsonDecorator(idx.POST)))
	mux.GET("/v1/api", printDecorator(jsonDecorator(idx.GET)))
	err := http.ListenAndServe(":900", mux)
	fmt.Printf("Closing with err: %v\n", err)

}
