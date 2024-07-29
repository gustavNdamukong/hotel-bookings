package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form creates a custom form struct and embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// Valid returns true if there are not errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initializes a form struct
// data will be empty when initialised
// NOTES: The way New() works works on a struct is to simulate the instantiation of objects in other langs,
// It works like a constructor accepting arguments to use to pass to the existing empty struct, thus initialising
// it with data. It returns a pointer to that struct.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	// NOTES: its very important to try to get the form field from the receiver struct Form,
	//& not the Form from the submitted request (thus, f.Get(field) & not r.Form.Get(field))
	x := f.Get(field)
	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// Required checks for required fields
// the '...' argument makes this a variatic function; meaning you can
// make this how many arguments as you want, & in the body u can loop/range
// through them & handle them any way you want
func (f *Form) Required(fields ...string) {
	//range through the supplied fields
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MinLength check for minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

func (f *Form) IsEmail(field string) bool {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
		return false
	}
	return true
}
