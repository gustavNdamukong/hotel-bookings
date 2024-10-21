package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gustavNdamukong/hotel-bookings/internal/config"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	// on a server we want as much info as possible on the error, so let's trace the error
	// NOTES: debug.stack() is how u get the stacktrace of an arror
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// NOTES: ideally we should email the site maintainer with a path to the error log file
	// but for now, let's just print the error to the terminal
	app.ErrorLog.Println(trace)
	// give some kind of feedback to the user
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// NOTES: This is how to quickly check if a user is logged in. Its simple-it returns true or false.
// Note that 'user_id' is the session var that is set when a user is successfully logged in.
// This will be called regularly by a middleware function Auth() (in 'cmd/web/middleware.go') in all
// pages to check if the user is logged in or not.
func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}
