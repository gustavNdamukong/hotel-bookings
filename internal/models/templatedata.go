package models

import "github.com/gustavNdamukong/hotel-bookings/internal/forms"

// TemplateData holds data to be sent to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32

	//if we aren't sure what the nature of the data will be, we make the value part an interface
	Data map[string]interface{}
	//cross-site-request protection
	CSRFToken string

	//temporal notification messages we may want to pass to the view files (flash, warning, or error messages)
	Flash   string
	Error   string
	Warning string

	//this will be used to validate forms on any page
	Form *forms.Form
}
