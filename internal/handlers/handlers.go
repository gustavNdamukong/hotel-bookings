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

	//----------------TESTING---------------------------
	/*viewError := m.App.Session.GetString(r.Context(), "error")
	stringMap := make(map[string]string)
	stringMap["error"] = viewError*/
	//----------------END TESTING-----------------------

	/////render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
	render.RenderTemplate(w, r, "index.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// send data to the template
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Majors suit page"

	// send the data to the template
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Search availability page"

	// send the data to the template
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	//this is how you extract values posted via a form
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	//cast the data to the required format
	// send the data to the template
	//convert the given text (to byte()) into a slice of bytes
	w.Write([]byte(fmt.Sprintf("start date is %s and end is %s", start, end)))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
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

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
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

	m.App.Session.Put(r.Context(), "reservation", reservation)
	//http response 'StatusSeeOther' is equal to http response code 303
	//which is ideal for redirections to handle post requests
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	//this is how you retrieve a struct passed to a session var. We use Session.Get(...)
	//and we chain the struct type at the end of it aka type-assert, thereby asserting that
	//what is stored in the 'reservation' key in the session is indeed a Reservation model.
	//Notice that this is as opposed to grabbing a string from the session-where you
	//would use Session.GetString()
	// TODO: document how you would store & retrieve a struct from the session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	/////remoteIP := m.App.Session.GetString(r.Context(), "remote_ip") /////
	/////log.Println("View data error is: ", reservation) /////
	if !ok {
		log.Println("can't get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		//the http response code 'StatusTemporaryRedirect' is essentially a 301 code
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// TODO: How to remove an item from the session
	m.App.Session.Remove(r.Context(), "reservation")

	// TODO: document in data types how when initialising a map, the interface{} type
	// declared for tits value indicates it will contain a struct. In the case below,
	// reservation is a struct. Also reference notes on how structs can be interfaces.
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
