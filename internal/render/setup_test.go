package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

/*
We need this setup test file because we are principally testing the 'AddDefaultData()' function
which used various dependencies like the session, config.AppConfig, http.Request etc
*/
var session *scs.SessionManager
var testApp config.AppConfig

// TestMain() is a standard function that is used in all setup_test.go files. It will be run
// first before any other tests are run. It should accept the 'm *testing.M' as its argument
// it will use 'testing.M.run()' (m.run()) at the end of this TestMain() function to continue
// to run all your other tests.
func TestMain(m *testing.M) {

	gob.Register(models.Reservation{})

	// change this to true when in production
	testApp.InProduction = false

	// set up logging. Create a new logger that writes to the terminal (os.Stdout), prefix the msg
	// with "INFO" & a tab, followed by the date & time
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

type myWriter struct{}

func (tw *myWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *myWriter) WriteHeader(i int) {} //no need to return anything

func (tw *myWriter) Write(b []byte) (int, error) {
	// NOTES: it is crucial here that you return the length (size) of byte
	length := len(b)
	return length, nil
}
