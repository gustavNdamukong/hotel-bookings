package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	// create an http handler (aka a MUX or multiplexer)
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf) //ignore any post request that doesn't have a proper CSRF token
	// NOTES: Here is how you use a middleware already defined in 'cmd/web/middleware.go/
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)

	// NOTES: How to parse a URL parameter sent from an HTML link
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)
	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)

	mux.Get("/user/logout", handlers.Repo.Logout)

	//create a file server to serve any files or images etc
	fileserver := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileserver))

	// NOTES: How to define a group of routes only available ONLY to authenticated users
	// In this case, we are saying this should apply to any route that starts with '/admin'. This will be
	// the group eg '/admin/properties', '/admin/dashboard' etc
	mux.Route("/admin", func(mux chi.Router) {
		// NOTES: Here is how you use a middleware. This middleware 'Auth' is defined in 'cmd/web/middleware.go/
		// in this case, we want to apply the 'Auth' middleware to all routes in this group, which inthis case will
		// only allow access to authenticaterd users.

		// NOTES: Commenting the following line out (mux.Use(Auth)) turns off authentrication for this route group
		/////mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminShowPostReservation)
	})

	return mux
}
