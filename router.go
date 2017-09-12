package easyroute

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"strconv"
	gobrake "gopkg.in/airbrake/gobrake.v2"
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

	airbrakeProjectId int64
	airbrakeProjectKey string
	airbrakeEnabled bool
}

// NewRouter creates a new easyroute Router object with the provided
// before handler and logger struct
func NewRouter(beforeFn beforeHandlerFunc, logger Logger, airbrakeInfo map[string]string) Router {
	muxRouter := mux.NewRouter()
	var enableAirbrake bool
	if(airbrakeInfo["enable"] == "true"){
		enableAirbrake = true
	} else {
		enableAirbrake = false
	}
	airbrakeId, err := strconv.Atoi(airbrakeInfo["projectId"])
	if(err != nil){
		enableAirbrake = false
	}
	router := Router{
		muxRouter,
		beforeFn,
		logger,
		int64(airbrakeId),
		airbrakeInfo["projectKey"],
		enableAirbrake,
	}

	return router
}

// SubRoute creates a new easyroute Router off the base router with provided
// prefix. This preserves the same before handler.
func (g *Router) SubRoute(prefix string) Router {
	muxRouter := g.PathPrefix(prefix).Subrouter()

	router := Router{
		muxRouter,
		g.beforeHandler,
		g.logger,
		0,
		"",
		false,
	}

	return router
}

// SubRouteC creates a new easyroute Router off the base router with provided
// prefix and an additional before handler.
// The routes in this router will now run through first the parent(base) router's
// before handler and then this router's before handler.
func (g *Router) SubRouteC(prefix string, beforeFn beforeHandlerFunc) Router {
	muxRouter := g.PathPrefix(prefix).Subrouter()

	router := Router{
		muxRouter,
		func(r *Request) bool {
			if g.beforeHandler(r) {
				return beforeFn(r)
			}
			return false
		},
		g.logger,
		0,
		"",
		false,
	}

	return router
}

func (g *Router) requestHandler(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if(g.airbrakeEnabled == true){
			airbrake := gobrake.NewNotifier(g.airbrakeProjectId, g.airbrakeProjectKey)
			defer airbrake.Close()
			defer airbrake.NotifyOnPanic()
		}
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
			userUuid := request.UserUuid
			request.Body(body)
			elapsed := time.Since(start)
			g.logger.LogI("origin=%s method=%s path=%s body=%s user_uuid=%s elapsed=%s", origin, method, path, body, userUuid, elapsed)
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
