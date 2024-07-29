package main

import (
	"fmt"
	"net/http"
	"testing"
)

/*
	NOTES:To test our app's middleware.go file which contains the two funcs:
		NoSurf() and
		SessionLoad()

	We know these two func both accept and return an http.Handler
	We need to find a way to make sure we get an http.Handler to be passed
	to them, and then test that the return data from them is actually an http.Handler as well.

	This is why we need a setup test file where we create a custom http.Handler that we can pass
	into them in here (myH as seen below)
*/

func TestNoSurf(t *testing.T) {
	// we need a way to setup the environment before this test runs
	var myH myHandler
	h := NoSurf(&myH)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not an http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	// we need a way to setup the environment before this test runs
	var myH myHandler
	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not an http.Handler, but is %T", v))
	}
}
