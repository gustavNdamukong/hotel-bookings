package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/forms"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// get the remote IP of the visitor (get that from the request obj)
	remoteIP := r.RemoteAddr
	log.Println("Remote IP detected: ", remoteIP)

	// every time someone hits that home page; get that user's IP addr as a string
	// & sSession.Put(r.Context(), "remote_ip", remoteIP) store it in the session with its session key being "remote_ip"
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	/////render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
	render.RenderTemplate(w, r, "index.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	// perform some logic
	stringMap := make(map[string]string)
	stringMap["title"] = "About us page"

	// pull value out of the session. Note that remoteIP will be an empty string
	// if there is nothing in the session named 'remote_ip'
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// send the data to the template
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// the handler for the generals page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {

	// perform some logic
	stringMap := make(map[string]string)
	stringMap["title"] = "Generals page"

	// send the data to the template
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// the handler for the majors page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {

	// perform some logic
	stringMap := make(map[string]string)
	stringMap["title"] = "Majors suit page"

	// send the data to the template
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// the handler for the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {

	// perform some logic
	stringMap := make(map[string]string)
	stringMap["title"] = "Search availability page"

	// send the data to the template
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// the handler for the search availability form submission
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	//this is how you extract values posted via a form
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	//cast the data to the required format
	// send the data to the template
	//convert the given text (to byte()) into a slice of bytes
	w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))
}

// convention in Go is to place any struct u write as close as possible to the code that uses it.
// in this case, the AvailabilityJSON() func need it.
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability & returns a JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      false,
		Message: "Available",
	}

	//convert the data into json. In Go we use Marshal which comes in many forms depending on how
	//you wanna use it-one of which is 'MarshalIndent()'
	output, err := json.MarshalIndent(resp, "", "     ") //indent by 5 spaces
	if err != nil {
		log.Println(err)
	}

	log.Println(string(output))

	//this is how you set request response headers, which in this case is critical
	//as we are returning json data. Do this before you write the actual response
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)

	//this is how you extract values posted via a form
	//start := r.Form.Get("start")
	//end := r.Form.Get("end")

	//cast the data to the required format
	// send the data to the template
	//convert the given text (to byte()) into a slice of bytes
	//w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))
}

// the handler for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {

	// perform some logic
	stringMap := make(map[string]string)
	stringMap["title"] = "Contact page"

	// send the data to the template
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// the handler for the reservation page page
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	stringMap := make(map[string]string)
	stringMap["title"] = "Reservation page"

	// send the data to the template
	//notice how we send an empty form to the target form view.
	//We initialise it with no data (nil) for submitted values
	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      forms.New(nil),
		Data:      data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	//use the form object & populate it with the submitted form data
	//Notice how to get form data from the request (request.Form.Get(fieldName))
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
	//validate submitted email address using the installed Govalidator library
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	//Note how we add values to sessions and how we redirect users to other pages
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

	// perform some logic
	/*stringMap := make(map[string]string)
	stringMap["title"] = "Reservation page"

	// send the data to the template
	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      form,
		//Data:      data,
	})

	return */
}

// ReservationSummary displays the res summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	//this is how we retrieve data from a session & pass it to another view file
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("Cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
