package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/handlers"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main function
func main() {
	// Register the models.Reservation type with gob
	// what am i going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	// initialise a session
	session = scs.New()

	// optionally set lifetime of session
	// 24 hours. A syntax error in this time specification will cause the session setting & retrieving of data not to work
	session.Lifetime = 24 * time.Hour

	// Name sets the name of the session cookie. It should not contain
	// The default cookie name is "session".
	// If your application uses two different sessions, you must make sure that
	// the cookie name for each of these sessions is unique.
	session.Cookie.Name = "testProj_session_id"
	//by default it uses cookie for itas data storage, but it has different storages u can choose from eg DBs
	session.Cookie.Persist = true // should the cookie persist after user closes the browser?
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction // set to true when using https in production

	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
	}

	app.DefaultAppTitle = "Hotel Reservation App"
	app.TemplateCache = templateCache

	//do a random global config setting change to test
	app.UseCache = false

	//set things up with our handlers
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	fmt.Printf(fmt.Sprintf("Starting application on port %s", portNumber))
	//_ = http.ListenAndServe(portNumber, nil)

	serve := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	log.Fatal(err)
}
