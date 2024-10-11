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

	//create a file server to serve any files or images etc
	fileserver := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileserver))

	return mux
}
