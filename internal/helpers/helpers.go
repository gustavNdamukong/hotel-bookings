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
