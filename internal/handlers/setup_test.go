package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates"
var functions = template.FuncMap{
	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// change this to true when in production
	app.InProduction = false

	// set up logging. Create a new logger that writes to the terminal (os.Stdout), prefix the msg
	// with "INFO" & a tab, followed by the date & time
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
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

	//to allow tests to pass, we need to simulate the sending and listening for mails which
	//is happening in our mail application, for we do not want to actually send emails here
	//during testing. Hence we create a similar mail channel here in the testing area with
	//the same kinda func listenForMail() which will do something else. This way we will
	//still achieve good test coverage.
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
	}

	app.TemplateCache = templateCache

	//do a random global config setting change to test
	app.UseCache = true

	//set things up with our handlers
	repo := NewTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func listenForMail() {
	//do the same thing that this same func does in the main app,
	//but avoid sending an email
	go func() {
		for {
			//do nopthing with what we get back from the MailChannel
			_ = <-app.MailChan
		}
	}()
}

func getRoutes() http.Handler {
	// what am I going to put in the session
	//we do this coz one of our handlers needs it
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	// set up logging. Create a new logger that writes to the terminal (os.Stdout), prefix the msg
	// with "INFO" & a tab, followed by the date & time
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up the session. We need to do it exactly as we do in the main app
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc

	// make app.UseCache true as we need to test with caching & the path to the
	// templates will be like 'pathToTemplates' above
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	// TODO: 'render.NewTemplates' has been changed to 'render.NewRenderer'. Confirm all is okay
	render.NewRenderer(&app)

	mux := chi.NewRouter()

	//test everything in routes.go
	mux.Use(middleware.Recoverer)

	// NOTES: note that you need to disable CSRFToken checks for test post requests or the tests will fail.
	// unless you try & pass in a CSRFToken with every test request, but testing that is not crucial,
	// so its best to disable that check (comment out 'NoSurf' which is the library we used to inforce that).
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/contact", Repo.Contact)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	//-----------------------------------
	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Get("/admin/dashboard", Repo.AdminDashboard)

	mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
	mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
	mux.Get("/admin/reservations-calendar", Repo.AdminReservationsCalendar)
	mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationsCalendar)
	mux.Get("/admin/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
	mux.Get("/admin/delete-reservation/{src}/{id}/do", Repo.AdminDeleteReservation)

	mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservation)
	mux.Post("/admin/reservations/{src}/{id}", Repo.AdminShowPostReservation)
	//-----------------------------------

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

// NoSurf is the csrf protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
