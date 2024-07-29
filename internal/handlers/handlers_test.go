package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"generals-quarters", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"majors-suite", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"search-availability", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"make-res", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-search-availability", "/search-availability", "Post", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"post-search-availability-json", "/search-availability-json", "Post", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"make-reservation", "/make-reservation", "Post", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "me@here.com"},
		{key: "phone", value: "555-555-5555"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	// NOTES: this is how you create test server in go which also provides you with a client
	// NewTLSServer starts and returns a new [Server] using TLS. The caller should call Close when finished, to shut it down
	ts := httptest.NewTLSServer(routes)

	// NOTES: defer does not close till the current function is terminated
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
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
		} else {
			// NOTES: url.Values{} is a built-in way to simulate and holds values for a post request
			// this will automaticlaly work for the test post request data prepared above-to be captured as e.params
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}

			// NOTES: this is how to create a clientwith the test server created above, to submit
			// requests to the given test URLs passing in the (post) data stored in 'values'
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				// fail the test
				t.Fatal(err)
			}

			// NOTES: check the test request response code, & see if it matches the expected test response code
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
