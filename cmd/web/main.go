package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/driver"
	"github.com/gustavNdamukong/hotel-bookings/internal/handlers"
	"github.com/gustavNdamukong/hotel-bookings/internal/helpers"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {
	// NOTES: add to debug notes that the equivalent of dump & die in go is
	// log.Fatal(err) coz it will abort the app execution & log the error. Remember to import log above though
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	fmt.Printf(fmt.Sprintf("Starting application on port %s", portNumber))

	serve := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// Register the models.Reservation type with gob
	// What akind of stuff will i be putting in the session. Register them all here
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// change this to true when in production
	app.InProduction = false

	// set up logging. Create a new logger that writes to the terminal (os.Stdout), prefix the msg
	// with "INFO" & a tab, followed by the date & time
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

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

	// connect to DB
	log.Println("Connecting to DB")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=hotel-bookings user=user password=")
	if err != nil {
		log.Fatal("Cannot cronnecting to database! Dying...")
	}
	log.Println("Connected to database")

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return nil, err
	}

	app.DefaultAppTitle = "Hotel Reservation App"
	app.TemplateCache = templateCache

	//do a random global config setting change to test
	app.UseCache = false

	//set things up with our handlers
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
