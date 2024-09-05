package easyroute

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	beforeHandler := func(req *Request) bool {
		return false
	}
	t.Run("basic creation and route registering", func(t *testing.T) {
		// setup a router and register a test route
		r := NewRouter(beforeHandler, Logger{}, "testServiceName")
		r.Get("/test", func(req *Request) {
			// empty handler
		})

		// check to make sure we registered the route with the internal mux.Router
		// r.Router --> muxtrace.Router
		// r.Router.Router --> mux.Router that is embedded in the muxtrace.Router
		paths := []string{}
		r.Router.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				t.Fatalf("Error getting path template: %s", err.Error())
			}
			paths = append(paths, path)
			return nil
		})

		assert.Contains(t, paths, "/test")
	})
}
