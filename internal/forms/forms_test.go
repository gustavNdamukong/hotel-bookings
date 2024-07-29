package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestForm_Valid NOTES: because Valid() is a receiver func for a 'Form' struct (in forms.go),
// the name of this test func to test Valid() will be 'TestForm_Valid' by convention.
func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

// NOTES: because Required() is a receiver func for a 'Form' struct (in forms.go),
// the name of this test func to test Required() will be 'TestForm_Required' by convention.
func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	// NOTES: here we are making fields required which have not been set
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing & it should fail coz they are required")
	}

	// NOTES: this is how u can simulate form values to submit to a URL either via POST or GET using url.Values{} (built-in).
	// Use Add() on the postedData to add more values to imaginary fields.
	// this is how you set the fields (the keys are synonymous to the field names)
	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	// NOTES: this is how you make the actual (in this case post) request to a URL. Use http.NewRequest().
	// we make a new request again here so we can get a fresh (separate) response from the one above
	// where we tested without setting fields
	// You can make a URL up eg "/whatever")
	r, _ = http.NewRequest("POST", "/whatever", nil)

	// NOTES: load the posted data onto the request response's 'PostForm' property & pass it to New()
	// so u can then simulate a handling (eg validation) of the submitted request
	r.PostForm = postedData // NOTES: We would have skipped this line if the form had no posted data
	form = New(r.PostForm)
	// NOTES: now its ok to make the fields required because they've been set above & given values (using postedData.Add())
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows does not have required fields when it does")
	}
}

func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	// NOTES: because we pass nothing to the post here
	// (like postedData := url.Values{} and postedData.Add("a", "a") and r.PostForm = postedData),
	// the following check should fail coz the submitted form has no field in it
	has := form.Has("whatever")
	if has {
		t.Error("Form shows field exists when it does not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)

	has = form.Has("a")
	if !has {
		t.Error("Form shows does not have field when field exists")
	}
}

func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	/* THIS WORKS
	// NOTES: We would have skipped the next 3 lines if the form had no posted data
	postedData := url.Values{}
	postedData.Add("a", "ab") //NOTES: make the value only 2 characters
	r.PostForm = postedData

	form := New(r.PostForm)

	// NOTES: here is how u get the name of a field added to a form (with form.Get(fieldName))
	// but you can still just say: form.MinLength("a", 3, r) since you know the field name you added was "a"
	testField := form.Get("a")
	hasMinLength := form.MinLength(testField, 3, r)
	if hasMinLength {
		t.Error("Form says field has at least 3 characters when it has only 2")
	} */

	// Alternative test code (no need to add fields and data to them)
	form := New(r.PostForm)
	form.MinLength("whateverField", 10)
	if form.Valid() {
		t.Error("Form shows min length for non-existent field")
	}

	// Check that the form throws validation errors when errors occur.
	// We do this here coz in the case of the "whateverField" above,
	// it does not exist so there should be an error
	isError := form.Errors.Get("whateverField")
	if isError == "" {
		t.Error("Should have an error but did no get one")
	}

	// Alternative test code (now lets add fields and data to them)
	//-------------------------------------
	postedValues := url.Values{}
	postedValues.Add("whateverField", "five") // make the field only 5 characters
	form = New(postedValues)

	form.MinLength("whateverField", 10)
	if form.Valid() {
		t.Error("Form shows field meets min length of 10 which is not")
	}

	// Let's also test that the min length works when the right length is submitted
	//-------------------------------------
	// NOTES: this is how you re-initialise post values to run multiple tests in same func
	// you re-assign url.Values{} to a variable using '=' not ':='
	postedValues = url.Values{}
	postedValues.Add("anotherField", "more than ten characters") // make the field more than characters
	form = New(postedValues)
	form.MinLength("anotherField", 10)
	if !form.Valid() {
		t.Error("Form shows field does not meet min length of 10 when it does")
	}

	// Check that the form throws no validation errors when no error occurs.
	// We do this here coz in the case of the "anotherField" above,
	// the field does exist (we added it above using 'postedValues.Add()') so there should be no error
	isError = form.Errors.Get("another_field")
	if isError != "" {
		t.Error("Should not have an error but got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	// Test for a) if field is blank, b) if wrong email is given & c) if correct email is given

	// a) if field is blank
	//----------------------------
	postedData := url.Values{}
	form := New(postedData) // NOTES: this new() requires url.Values it does not require a request

	form.IsEmail("someField") // field does not exist
	if form.Valid() {
		t.Error("Form sahows valid email for non-existent field")
	}

	// a) if wrong email is given
	//----------------------------
	// NOTES: We would have skipped the next 3 lines if the form had no posted data
	postedData = url.Values{}
	postedData.Add("email", "ab@gus") // NOTES: pass wrong email

	form = New(postedData)

	emailField := form.Get("email")
	isEmail := form.IsEmail(emailField)

	if isEmail {
		t.Error("Form says email field is valid when it is not")
	}

	// a) if correct email is given
	//----------------------------
	postedData = url.Values{}
	postedData.Add("email", "ab@gus.com") // NOTES: pass correct email

	form = New(postedData)

	isEmail = form.IsEmail("email")

	if !isEmail {
		t.Error("Form says email field is invalid when it is valid")
	}
}
