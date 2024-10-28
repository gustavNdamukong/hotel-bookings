package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/driver"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

type postData struct {
	key   string
	value string
}

// theTests is a slice of structs. Here we follow the table test approach
// Notice that each struct passed in here is for each view page of our app.
// The 'url' will be the route to access this view &
// the 'method' will be the http method to send a request to that route.
// Here we are testing our views by sending requests to them & checking
// that the returned http status code is as expected.
// To use this slice, wherever you call it from, you would do so by ranging over
// over it, which means you would be ranging over the structs here, while doing
// something with each. The values in the opening curly braces are the keys for
// the properties each struct should have, and their data types.
var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"generals-quarters", "/generals-quarters", "GET", http.StatusOK},
	{"majors-suite", "/majors-suite", "GET", http.StatusOK},
	{"search-availability", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	{"show res cal", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"show res cal with params", "/admin/reservations-calendar?y=2020&m=1", "GET", http.StatusOK},

	// {"post-search-availability", "/search-availability", "Post", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"post-search-availability-json", "/search-availability-json", "Post", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"make-reservation", "/make-reservation", "Post", []postData{
	// 	{key: "first_name", value: "John"},
	// 	{key: "last_name", value: "Smith"},
	// 	{key: "email", value: "me@here.com"},
	// 	{key: "phone", value: "555-555-5555"},
	// }, http.StatusOK},
	{"non-existent", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	// NOTES: this is how you create test server in go which also provides you with a client
	// NewTLSServer starts and returns a new [Server] using TLS. The caller should call Close when finished, to shut it down
	ts := httptest.NewTLSServer(routes)

	// NOTES: defer does not close till the current function is terminated
	defer ts.Close()

	for _, e := range theTests {
		// create a web client to make requests to our test web server
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			// NOTES: This is how you fail a test in go
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			// NOTES: error handling, this is how we can generate a test error in a formatted string that
			// can take variables. t.Error will only generate an error string (not formatted)
			t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("Get", "/make-reservation", nil)
	//get context to pass with our request
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	//now we need an httptest.NewRecorder (built into Go) which fakes the entire process of sending a request & getting a response back
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	//call our reservation function (route handler)
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	//check if test passed
	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, instead of %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	//check if test passed
	/*
		NOTES:
		Redirect code meanings
			-http.StatusSeeOther 			= 307 (the feature being tested literally redirected with 'StatusSeeOther')
			-http.StatusOK 					= 200 (when feature works ok eg a page redirecting to the next view implies code 200 )
			-http.StatusTemporaryRedirect	= 303
	*/
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, instead of %d", rr.Code, http.StatusSeeOther)
	}

	// test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomId = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	//check if test passed
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, instead of %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	// NOTES: This is how to build the body for the post request
	//	The 3rd argument to NewRequest() below for POST requests must be an IO reader.
	//	the strings function of the standard library has one as a method called NewReader()
	//	to which you need to pass the POST body string (in our case reqBody)
	//-----------------------------------------------------------------------------
	/*
		reqBody := "start_date=2050-01-01"
		reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
		reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
		reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
		reqBody = fmt.Sprintf("%s&%s", reqBody, "email=2john@smith.ca")
		reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=205-123-1234")
		reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	*/
	//-----------------------------------------------------------------------------

	//------------------Alternative way to emulate post request--------------------
	// NOTES: Here's a mush easier way than the above method, to build data for a post request
	//	You can replace reqBody above with this:
	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "555-555-5555")
	postedData.Add("room_id", "1")
	//------------------End post request alternative-------------------------------

	//req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// set the request header (good practice to always do so)
	// NOTES: Here is how you set a request header. Note that the value of "Content-Type" must always be
	//	"application/x-www-form-urlencoded" for post requests. That is how you tell the receiving server
	//	you are transmitting data to that the data being sent through is a post/form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// get your handler-which in our case is the PostReservation() method in '/internal/handlers/handler.go'
	handler := http.HandlerFunc(Repo.PostReservation)
	// call gthe handler using ServeHTTP() passing it the request writer (rr) and the actual request (re)
	//	Note that in this testing scenario, request writer (rr) is not an actual request writer but we built it to
	//		satisfy the requirements of being a request writer.
	handler.ServeHTTP(rr, req)

	//check if test passed
	// Note that we check for an http return code here of 'StatusSeeOther' because that's what the handler for this request
	//	(PostReservation() in '/internal/handlers/handler.go') returns
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, instead of %d", rr.Code, http.StatusSeeOther)
	}

	// test for missing post body
	//NOTES: How to test a post request. Notice the crucial 'Content-Type' header param
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for invalid start date
	postedData = url.Values{}
	postedData.Add("start_date", "invalid")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid start date: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for invalid end date
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "invalid")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid end date: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for invalid room id
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "invalid")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid room id: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for invalid data
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "J")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	/////if rr.Code != http.StatusSeeOther {
	if rr.Code != http.StatusOK {
		/////t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusSeeOther)
		t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test for failure to insert reservation into database
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "2")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler failed when trying to insert reservation: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for failure to insert restriction into database
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.ca")
	postedData.Add("phone", "1234567890")
	postedData.Add("room_id", "1000")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler failed when trying to insert restriction: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	/*****************************************
	// first case -- rooms are not available
	*****************************************/
	// create our request body
	reqBody := "start=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")

	// create our request
	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))

	// get the context with session
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr := httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler := http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have no rooms available, we expect to get status http.StatusSeeOther
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability when no rooms are available gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// second case -- rooms are available
	*****************************************/
	// this time, we specify a start date before 2050-01-01, which will give us
	// a non-empty slice, indicating that rooms are available
	reqBody = "start=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")

	// create our request
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have rooms available, we expect to get status http.StatusOK
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability when rooms are available gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// third case -- empty post body
	*****************************************/
	// create our request with a nil body, so parsing form fails
	req, _ = http.NewRequest("POST", "/search-availability", nil)

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have rooms available, we expect to get status http.StatusSeeOther
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability with empty request body (nil) gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// fourth case -- start date in wrong format
	*****************************************/
	// this time, we specify a start date in the wrong format
	reqBody = "start=invalid"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2040-01-02")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have rooms available, we expect to get status http.StatusSeeOther
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability with invalid start date gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// fifth case -- end date in wrong format
	*****************************************/
	// this time, we specify a start date in the wrong format
	reqBody = "start=2040-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "invalid")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have rooms available, we expect to get status http.StatusSeeOther
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability with invalid end date gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// sixth case -- database query fails
	*****************************************/
	// this time, we specify a start date of 2060-01-01, which will cause
	// our testdb repo to return an error
	reqBody = "start=2060-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-01-02")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.PostAvailability)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have rooms available, we expect to get status http.StatusSeeOther
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Post availability when database query fails gave wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	/*****************************************
	// first case -- rooms are not available
	*****************************************/
	/*// create our request body
	// To ensure no room is found available, we do not pass a room_id with the request
	reqBody := "start=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")

	// create our request
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	// get the context with session
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr := httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have no rooms available, we expect to get status http.StatusSeeOther
	// this time we want to parse JSON and get the expected response
	var j jsonResponse
	err := json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// Now we just nead into the response we are getting back above and complete the test
	// since we specified a start date > 2049-12-31, we expect no availability
	if j.OK {
		// NOTES: How we throw an error to fail a test (using t.Error())
		t.Error("Got availability when none was expected in AvailabilityJSON")
	}
	*/
	/*****************************************
	// second case -- rooms are available
	*****************************************/
	/*// create our request body
	reqBody = "start=2040-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2040-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	// create our request
	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if !j.OK {
		t.Error("Got no availability when some was expected in AvailabilityJSON")
	}
	*/
	/*****************************************
	// third case -- no request body
	*****************************************/
	/*// create our request
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if j.OK || j.Message != "Internal server error" {
		t.Error("Got availability when request body was empty")
	}
	*/
	/*****************************************
	// fourth case -- database error
	*****************************************/
	/*// create our request body
	reqBody = "start=2060-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	// get the context with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	rr = httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if j.OK || j.Message != "Error querying database" {
		t.Error("Got availability when simulating database error")
	}
	*/

	//------------------------------------------------------------------------------------------
	// NOTES:
	// BELOW IS AN ALTERNATIVE WAY TO RUN ALL THE TESTS ABOVE IN THIS 'AvailabilityJSON()'
	//	METHOD. IT IS A TABLE TEST, WHICH USES A SLICE OF STRUCTS EACH OF WHICH CONTAINS
	//	DATA FOR A TEST REQUEST. NOTE HOW WE USE THE SLICE OF STRUCTS (testAvailabilityJSONData)
	//	BY LOOPING THROUGH IT AND MAKING THE TEST REQUESTS WITH THE DATA OF EACH STRUCT. THIS
	//	TABLE TEST APPROACH RESULTS IN A STREAM-LINED AND READABLE TEST CODE.
	//------------------------------------------------------------------------------------------
	for _, e := range testAvailabilityJSONData {
		// create request, get the context with session, set header, create recorder
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal([]byte(rr.Body.String()), &j)
		if err != nil {
			t.Error("failed to parse json!")
		}

		if j.OK != e.expectedOK {
			t.Errorf("%s: expected %v but got %v", e.name, e.expectedOK, j.OK)
		}
	}
}

// testAvailabilityJSONData is data for the AvailabilityJSON handler, /search-availability-json route
var testAvailabilityJSONData = []struct {
	name            string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start":   {"2050-01-01"},
			"end":     {"2050-01-02"},
			"room_id": {"1"},
		},
		expectedOK: false,
	}, {
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedOK: true,
	},
	{
		name:            "empty post body",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "Internal Server Error",
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start":   {"2060-01-01"},
			"end":     {"2060-01-02"},
			"room_id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Error querying database",
	},
}

// ------------------------------------------------------------------------------------------
// NOTES:
// BELOW IS AN ALTERNATIVE WAY WE COULD HAVE TESTED THE 'PostAvailability()' METHOD ABOVE WITH
//	ALL ITS TEST SCENARIOS:
//		-rooms not available
//		-rooms are available
//		-empty post body
//		-start_date wrong format
//		-end_date wrong format
//		-database query fails
//
// IN THE Availability() METHOD, WE USE THE DATA COMING FROM THE SLICE OF STRUCTS IN
// TestRepository_PostAvailability()
// ------------------------------------------------------------------------------------------
// TestPostAvailability tests the PostAvailabilityHandler
/*func TestPostAvailability(t *testing.T) {
	for _, e := range testPostAvailabilityData {
		req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(e.postedData.Encode()))

		// get the context with session
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the request header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

// testPostAvailabilityData is data for the PostAvailability handler test, /search-availability
var testPostAvailabilityData = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start": {"2050-01-01"},
			"end":   {"2050-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "empty post body",
		postedData:         url.Values{},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "start date wrong format",
		postedData: url.Values{
			"start":   {"invalid"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "end date wrong format",
		postedData: url.Values{
			"start": {"2040-01-01"},
			"end":   {"invalid"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start": {"2060-01-01"},
			"end":   {"2060-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}
*/

func TestRepository_ReservationSummary(t *testing.T) {
	/*****************************************
	// first case -- reservation in session
	*****************************************/
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	/*****************************************
	// second case -- reservation not in session
	*****************************************/
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	/*****************************************
	// first case -- reservation in session
	*****************************************/
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	// set the RequestURI on the request so that we can grab the ID
	// from the URL
	req.RequestURI = "/choose-room/1"

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	///*****************************************
	// second case -- reservation not in session
	//*****************************************/
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	///*****************************************
	// third case -- missing url parameter, or malformed parameter
	//*****************************************/
	req, _ = http.NewRequest("GET", "/choose-room/fish", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/fish"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	/*****************************************
	// first case -- database works
	*****************************************/
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/book-room?s=2050-01-01&e=2050-01-02&id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	/*****************************************
	// second case -- database failed
	*****************************************/
	req, _ = http.NewRequest("GET", "/book-room?s=2040-01-01&e=2040-01-02&id=4", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

// loginTests is the data for the Login handler tests
var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"me@here.ca",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"jack@nimble.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"j",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func TestLogin(t *testing.T) {
	// range through all tests
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		// create request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the header (good practice)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// checking for expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostShowReservationTests = []struct {
	name                 string
	url                  string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data-from-new",
		url:  "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-new",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-all",
		url:  "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-all",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-cal",
		url:  "/admin/reservations/cal/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2022"},
			"month":      {"01"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-calendar?y=2022&m=01",
		expectedHTML:         "",
	},
}

// TestAdminPostShowReservation tests the AdminPostReservation handler
func TestAdminPostShowReservation(t *testing.T) {
	for _, e := range adminPostShowReservationTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/user/login", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = e.url

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		///// handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler := http.HandlerFunc(Repo.AdminShowPostReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// checking for expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

func TestPostReservationCalendar(t *testing.T) {
	for _, e := range adminPostReservationCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

	}
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestAdminProcessReservation(t *testing.T) {
	for _, e := range adminProcessReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "delete-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "delete-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestAdminDeleteReservation(t *testing.T) {
	for _, e := range adminDeleteReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
