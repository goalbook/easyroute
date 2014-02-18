package easyroute

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type handlerFunc func(*Request)
type beforeHandlerFunc func(*Request) bool
type loggerFunc func(string, ...interface{})

type Logger struct {
	LogI loggerFunc
	LogE loggerFunc
	LogD loggerFunc
}

type Router struct {
	// Inherit a mux router
	*mux.Router

	beforeHandler beforeHandlerFunc
	logger        Logger
}

func NewRouter(beforeFn beforeHandlerFunc, logger Logger) Router {
	muxRouter := mux.NewRouter()

	gbrouter := Router{
		muxRouter,
		beforeFn,
		logger,
	}

	return gbrouter
}

func (g *Router) requestHandler(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body interface{}
		// Start timer
		start := time.Now()

		request := NewRequest(w, r)
		// Run Before block
		if g.beforeHandler(&request) == true {
			// If the before block returns false we don't execute the rest
			// Run actual handler
			fn(&request)
		}

		if g.logger.LogI != nil {
			// Log out some of the info
			origin := r.RemoteAddr
			method := r.Method
			path := r.URL.Path
			request.Body(body)
			elapsed := time.Since(start)
			g.logger.LogI("origin=%s method=%s path=%s body=%s elapsed=%s", origin, method, path, body, elapsed)
		}
	}
}

func (g *Router) Get(path string, handler handlerFunc) {
	g.HandleFunc(path, g.requestHandler(handler)).Methods("GET")
}

func (g *Router) Put(path string, handler handlerFunc) {
	g.HandleFunc(path, g.requestHandler(handler)).Methods("PUT")
}

func (g *Router) Post(path string, handler handlerFunc) {
	g.HandleFunc(path, g.requestHandler(handler)).Methods("POST")
}

func (g *Router) Delete(path string, handler handlerFunc) {
	g.HandleFunc(path, g.requestHandler(handler)).Methods("DELETE")
}
