package main

import (
	"fmt"
	"testing"

	"github.com/go-chi/chi"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
)

// TestRoutes needs to test that our routes() function (in /cmd/web/routes.go)
// returns a an instance of (pointer to) the chi router (chi.Mux)
func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing; test passed
	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux, type is %T", v))
	}
}
