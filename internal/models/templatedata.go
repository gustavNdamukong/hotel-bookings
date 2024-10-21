package models

import "github.com/gustavNdamukong/hotel-bookings/internal/forms"

// TemplateData holds data to be sent to templates/view files
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

	// NOTES: Since this struct is what carries data from the backend to all views, it makes sense to pass in
	// it a flag that will be used by views to know if a user is logged in. The base template file eg will
	// check it to know whether to show a login or logout link. If IsAuthenticated > 0 then user is logged in
	// but if IsAuthenticated == 0, then the user is logged out.We will therefore update this in the backend
	//whenever we login/logut a user.
	IsAuthenticated int
}
