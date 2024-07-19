package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf is the csrf protection middleware. It adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	// we need to store the generated csrf token so that it is available to be checked against after the form submission
	// this is also how you set a cookie in Go
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad a middleware func to make your application session-aware, in other words, make it use sessions
// Without it basically; you won't be able to save & retrieve data from a session
func SessionLoad(next http.Handler) http.Handler {
	// LoadAndSave() is a built-in func that auto-loads & saves session data for the current request &
	// sends the session token to & from the client in a cookie
	return session.LoadAndSave(next)
}
