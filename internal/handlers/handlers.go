package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/driver"
	"github.com/gustavNdamukong/hotel-bookings/internal/forms"
	"github.com/gustavNdamukong/hotel-bookings/internal/helpers"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
	"github.com/gustavNdamukong/hotel-bookings/internal/repository"
	"github.com/gustavNdamukong/hotel-bookings/internal/repository/dbrepo"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "index.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)

	// send data to the template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Majors suit page"

	// send the data to the template
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Search availability page"

	// send the data to the template
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	//this is how you extract values posted via a form
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// check availability of all rooms (it should return a slice of room models)
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// no available rooms
		m.App.Session.Put(r.Context(), "error", "No available room")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	/////w.Write([]byte(fmt.Sprintln("WHAT IS GOING ON???")))

	//prepare to send data of available rooms to the view to display to user
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// we need specific data about the dates (start & end) the user just choose to make a reservation for,
	// stored in a session for use in the template where we will be showing the user the rooms that are available
	// for booking. The available rooms will be displayed to them as links, and we will automatically apply the
	// users chosen reservation dates on which ever room they then choose.
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})

	//cast the data to the required format
	// send the data to the template
	//convert the given text (to byte()) into a slice of bytes
	//w.Write([]byte(fmt.Sprintf("start date is %s and end is %s", start, end)))
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	startD := r.Form.Get("start")
	endD := r.Form.Get("end")

	// NOTES: As usual, convert a date coming from the browser (a form)
	//	from a string to a Go time.Time which is expected by our custom
	//	'SearchAvailabilityByDatesByRoomId()' function in '/internal/repository/dbrepo/postgres.go'
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startD)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, endD)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// NOTES: Leaving out the Message key will have it default to a blank string anyway,
	//	though we just give it a blank string
	//	Note how for dates (start & end dates), we use the string versions coz we're passing these back to the view
	//	Note how we use strconv.Itoa this time (& not strconv.Atoi()) to convert the room ID from an int var to a
	//	struct string
	resp := jsonResponse{
		OK:        available,
		Message:   "",
		RoomID:    strconv.Itoa(roomID),
		StartDate: startD,
		EndDate:   endD,
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

	// NOTES: this is how you extract values posted via a form
	// start := r.Form.Get("start")
	// end := r.Form.Get("end")

	// If you were to print that data directly to the browser, you would need to cast it
	// to the required format as you write it to the browser like so:
	// converting the given text to (byte()) a slice of bytes
	// w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// Reservation renders the 'make-reservation' page and displays a form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// NOTES: How to generate an error string
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomById(reservation.RoomId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.Room.RoomName = room.RoomName

	// put the reservation model in the session as we will need it later
	m.App.Session.Put(r.Context(), "reservation", reservation)

	// NOTES: When it comes to dates, when accepting data with dates from the browser eg forms,
	//	the dates need to be converted from strings to time.Time, and vice versa. In this case,
	//	we need to pass data from the reservation model stored in the session to a view HTML form,
	//	so we need to re-convert the start & end dates to string formats again. Here is how to do it.
	startD := reservation.StartDate.Format("2006-01-02")
	endD := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = startD
	stringMap["end_date"] = endD
	stringMap["title"] = "Make Reservation"

	data := make(map[string]interface{})
	data["reservation"] = reservation

	// send the data to the template
	//notice how we send an empty form to the target form view.
	//We initialise it with no data (nil) for submitted values
	// NOTES: This is how you pass data to a view. In this case we pass data we have
	//	prepared above (stringMap & data) to the correspondiong keys: StringMap & Data
	//	that are part of the models.TemplateData struct that are passed to all views.
	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Form:      forms.New(nil),
		Data:      data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation data from session"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		// NOTES: this is how we can make use of our error helper to throw errors
		helpers.ServerError(w, err)
		return
	}

	/* NOTES - HANDLING DATES IN GO: We are receiving our 'start_date' and 'end_date' as
	//		strings (from the browser form, eg 2020-01-01)
	//		and we need to convert it into a format that our model expects.

	The conversion is a little tricky because of how Go handles the conversion of date strings-to-time objects,
	and the formatting of time objects to strings. Unlike other programming languages, Go uses an unusual format.


			Mon Jan 2 15:04:05 MST 2006 (MST aka Mountain Standard Time is GMT-0700)
			OR (in other words, in Golang, the reference data/time are
				Monday Jan 2nd at 4 minutes and 5 seconds after 3pm Mountain Standard time in the year 2006)

				This is based on the standard of writing dates in the USA, which is:

			01/02 03:04:05PM '06 -0700

				'06 stands for 2006, while -0700 stands for GMT, which represents Mountain Standard time

	Instead of having to remember or lookup the traditional formatting codes for functions like 'strftime', you just
	count 1234 while knowing that each place in the standard time corresponds to a component of a date/time object (Time type in Go).
	One will stand for day of the month, two for the month, three for the hour (in 12 hour clock), four for the mnutes etc
	(see this blog: https://www.pauladamsmith.com/bog/2011/05/go_time.html)

	The way you put this in action is by putting away the parts of the standard time in a layout string that mstches the format of
	either
		-the string representation of time you wish to parse into a Time object,
		or the opposite direction-generate a string representation of time from a Time object

	Feel free to look up the format each time you want to code/convert a date in Go. Basically, describe to Go the format in which
	your date from a form will be coming in, using this date format as a reference:

		2024-01-01 -- 01/02 03:04:05PM '06 -0700

	*/

	// Now let's use the form object & populate it with the submitted form data
	// Notice how we get the form field data from the request using request.Form.Get(fieldName)
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	// TODO: validate submitted email address using the installed Govalidator library
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Now save this reservation to the DB
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// prepare a model which we will send to a func to resstrict a room that has been reserved
	// that's why we pass it the newReservationID of the newly inserted reservation ID above
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationID: newReservationID,
		RestrictionID: 1,
		// These last 3 we might need in future, but tey're not part of the DB fields
		/*Room: 		xxxxx,
		Reservation:    xxxxx,
		Restriction:	xxxxx, */
	}

	//send the model to the func to do the restriction insertion
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)
	//http response 'StatusSeeOther' is equal to http response code 303
	//which is ideal for redirections to handle post requests
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	//NOTES: this is how you retrieve a struct passed to a session var. We use Session.Get(...)
	//and we chain the struct type at the end of it aka type-assert, thereby asserting that
	//what is stored in the 'reservation' key in the session is indeed a Reservation model.
	//Notice that this is as opposed to grabbing a string from the session-where you
	//would use Session.GetString()
	// NOTES: document how you would store & retrieve a struct from the session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Cannot get reservation from session")
		//the http response code 'StatusTemporaryRedirect' is essentially a 301 code
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// NOTES: How to remove an item from the session
	m.App.Session.Remove(r.Context(), "reservation")

	// NOTES: document in data types how when initialising a map, if an interface{} type
	// is declared for its value, that indicates it will contain a struct. In the case below,
	// reservation is a struct. Also refer to notes on how structs can be interfaces.
	data := make(map[string]interface{})
	data["reservation"] = reservation

	startD := reservation.StartDate.Format("2006-01-02")
	endD := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = startD
	stringMap["end_date"] = endD

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// NOTES: This is how to read URL params. A link from the view (choose-room.page.tmple)
	// 	sends an id to the route 'mux.Get("/choose-room/{{id}}", handlers.Repo.ChooseRoom)'
	//	which routes to this handle func, hence we need to retrieve the "id" URL param below.
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// let's pull the reservation credentials we put in the session earlier
	//  in this in case it was a model we put in
	//NOTES: This is how to retrieve data from a session. The values are usually strings by default.
	//	if it was not a string you put in the session, you must cast it to a string if you want to
	//  get a string back out like so: res, ok := m.App.Session.Get(r.Context(), "foo").(string).
	//  this will retrieve the session data "foo" as a string
	//	But in our case here, coz it was a Reservation model struct we put in the session, we need
	//  to cast it to a model, a reservation model to get the model back out, hence the line below:
	//  res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	// the reservation model we put in the session earlier had only the start & end date the user wanted
	// to book for. When they then choose which room the wanted, it brought them here where we now retrieve
	// the reservation model from the session, add the id of their chosen room, & then put the reservation
	// model back into the session for forwarding to the next screen which will be some kind of
	// 'make reservation page'.
	res.RoomId = roomID
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}
