package main

import (
	"net/http"

	"github.com/gustavNdamukong/hotel-bookings/internal/helpers"
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

// NOTES: SessionLoad is a middleware func to make your application session-aware, in other words, make it use sessions
// Without it basically; you won't be able to save & retrieve data from a session
func SessionLoad(next http.Handler) http.Handler {
	// LoadAndSave() is a built-in func that auto-loads & saves session data for the current request &
	// sends the session token to & from the client in a cookie
	return session.LoadAndSave(next)
}

// NOTES: Here is how you create a middleware. In this case we want to create one that will be used on
// all pages where we have access to this middleware and the helper file '/internal/helpers/helpers.go'
// which contains the 'IsAuthenticated() ' function to constantly check if a user is authenticated
// Note that this is another middleware function like the other ones in this file (NoSurf() and SessionLoad())
// which both accept the http.Handler, but the difference it that we make it use http.HandlerFunc() to call
// an anonymous function (like so: func(w http.ResponseWriter, r *http.Request)) which accesses the http request
// object and the calls our custom function IsAuthenticated() to check if the current user is loggen in
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
