package main

import "testing"

// NOTES: add note that to run tests from CLI, u need to be in directory where test files are
// if you want to run specific tests.
// NOTES: find out what you want to do if you want to run all tests in the app, in one tun
func TestRun(t *testing.T) {
	err := run()
	if err != nil {
		t.Error("Failed run()")
	}
}
