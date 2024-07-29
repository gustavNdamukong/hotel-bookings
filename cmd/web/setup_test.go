package main

import (
	"net/http"
	"os"
	"testing"
)

// TestMain() is a standard function that is used in all setup_test.go files. It will be run
// first before any other tests are run. It should accept the 'm *testing.M' as its argument
// it will use 'testing.M.run()' (m.run()) at the end of this TestMain() function to continue
// to run all your other tests.
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

// NOTES: this setup_test.go file will be run before any other tests are run.
// This is how you can simulate the built-in http.Handler of go so that you can pass it to test
// functions in your app that need one as an argument. You create a struct & create a receiver
// function for it that accepts these two things; a http.ResponseWriter & a http.Request as its arguments
// Here is a place to initialise variables that you may need in your application's testing
type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// no need to return anything
}
