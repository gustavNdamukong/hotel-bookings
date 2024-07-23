package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gustavNdamukong/hotel-bookings/pkg/config"
	"github.com/gustavNdamukong/hotel-bookings/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	// create an http handler (aka a MUX or multiplexer)
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals", handlers.Repo.Generals)
	mux.Get("/majors", handlers.Repo.Majors)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/reservation", handlers.Repo.Reservation)

	//create a file server to serve any files or images etc
	fileserver := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileserver))

	return mux
}
