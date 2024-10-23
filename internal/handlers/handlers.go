package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

// NewRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
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

// PostAvailability renders the search availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//this is how you extract values posted via a form
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// check availability of all rooms (it should return a slice of room models)
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		// no availabile rooms
		m.App.Session.Put(r.Context(), "error", "No availabile room")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	//prepare to send data of available rooms to the view to display to user
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// we need specific data about the dates (start & end) the user just chose to make a reservation for,
	// stored in a session for use in the template where we will be showing the user the rooms that are available
	// for booking. The available rooms will be displayed to them as links, and we will automatically apply the
	// users chosen reservation dates on whichever room they then choose.
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
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
	// need to parse request body (form) else we wont be able to write a test for this
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

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
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
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

	// I removed the error check, since we handle all aspects of the json right here
	//	as in, we are manually creating the json code above (resp) & therefore we know its contensts
	//	there will never be a situation where there will be an error
	out, _ := json.MarshalIndent(resp, "", "     ")

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
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		//NOTES: How to redirect user to another route
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomById(reservation.RoomId)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find room with that id")
		//NOTES: How to redirect user to another route
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
	/*
		reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
		if !ok {
			// NOTES: How to generate an error string (ServerError() is a custom function,
			//	checkout its content-in internal/helpers/helpers.go)
			helpers.ServerError(w, errors.New("Cannot get reservation data from session"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			m.App.Session.Put(r.Context(), "error", "cannot parse form")
			//NOTES: How to redirect user to another route
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	*/

	//-----------------------NOTES START-----------------------------------------------//
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
	//-----------------------NOTES END-----------------------------------------------//

	/*
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
			http.Error(w, "my own error message", http.StatusSeeOther)
			render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
				Form: form,
				Data: data,
			})
			return
		}

		// Now save this reservation to the DB
		newReservationID, err := m.DB.InsertReservation(reservation)
		if err != nil {
			m.App.Session.Put(r.Context(), "error", "cannot insert reservation into database")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			// NOTES: this is how we can make use of our error helper to throw errors
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
			Room: 		xxxxx,
			Reservation:    xxxxx,
			Restriction:	xxxxx,
		}


		//send the model to the func to do the restriction insertion
		err = m.DB.InsertRoomRestriction(restriction)
		if err != nil {
			m.App.Session.Put(r.Context(), "error", "cannot insert room restriction!")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		m.App.Session.Put(r.Context(), "reservation", reservation)
		//http response 'StatusSeeOther' is equal to http response code 303
		//which is ideal for redirections to handle post requests
		http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
	*/

	//-------------------------- TREVOR START-------------------------------

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse form")
		//NOTES: How to redirect user to another route
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	startD := r.Form.Get("start_date")
	endD := r.Form.Get("end_date")

	// 2020-01-01 -- 01/02 03:04:05PM '06 -0700

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, startD)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, endD)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomId:    roomID,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	// TODO: validate submitted email address using the installed Govalidator library
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "my own error message", http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Now save this reservation to the DB
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot insert reservation into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// NOTES: this is how we can make use of our error helper to throw errors
		return
	}

	// prepare a model which we will send to a func to resstrict a room that has been reserved
	// that's why we pass it the newReservationID of the newly inserted reservation ID above
	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomId:        roomID,
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
		m.App.Session.Put(r.Context(), "error", "cannot insert room restriction!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//-------------------------------------------
	// send email notifications - first to guest
	htmlMessage := fmt.Sprintf(`
			<strong>Reservation Confirmation</strong><br>
			Dear %s, <br>
			This is to confirm your reservation from %s to %s.
		`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "gustavfn@yahoo.co.uk",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg
	//-------------------------------------------

	//-------------------------------------------
	// send email notifications - first to property owner
	htmlMessage = fmt.Sprintf(`
			<strong>Reservation Notification</strong><br>
			Dear %s, <br>
			This is to notify you of a new reservation that has been booked for your property%s.<br>
			The Booking is by %s %s and the reservation is <br>
			from %s to %s.<br>

			Kind regards<br>
			The dream team
		`, "IDoNotKnowOwnerName", reservation.Room.RoomName, reservation.FirstName, reservation.LastName,
		reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:       "IDoNotKnowOwnerEmail@gmail.com",
		From:     "gustavfn@yahoo.co.uk",
		Subject:  "Reservation Notification",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg
	//-------------------------------------------

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

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// NOTES: This is how to read URL params. A link from the view (choose-room.page.tmple)
	// 	sends an id to the route 'mux.Get("/choose-room/{{id}}", handlers.Repo.ChooseRoom)'
	//	which routes to this handle func, hence we need to retrieve the "id" URL param below.

	// NOTES: This works well, however, this convenience function offered by chi, chi.URLPara(r, "id")
	//	is really, hard to test. In truth, we don't even need to use it, since we can parse the URL
	//	and find the id on our own. So change this code:

	/* roomID, err := strconv.Atoi(chi.URLParam(r, "id")) */

	// to this code: so we can test it more easily. Basically; we split the URL up by /, and grab the 3rd element
	// In your test for ChooseRoom(), you will want to set the URL on your request as follows: req.RequestURI = "/choose-room/1"
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

// BookRoom takes URL params, builds a sessional variable, and takes user to make-reservation screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// NOTES: Here is how you retrieve data from URL parameters (GET values)
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startD := r.URL.Query().Get("s")
	endD := r.URL.Query().Get("e")
	/////log.Println(ID, endDate, startDate)

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

	room, err := m.DB.GetRoomById(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var res models.Reservation
	res.RoomId = roomID
	res.Room.RoomName = room.RoomName
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "make-reservation", http.StatusSeeOther)

}

// ShowLogin shows the login screen
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// NOTES: For security reasons, renew the session token whenever you are doing a login or logout
	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)

	// validate form fields
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}

	//need to store their id in the session
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
// NOTES: Here is how you log a user out. Destroy the whole session & dont forget to renew the session token
// which you should do every time you log a user in or out.
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminReservations shows all reservations in admin dashboard
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations shows all new reservations in admin dashboard
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation in the admin dashboard
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	// NOTES: How to get a param value from the URL parameters
	// we need two bits from the URL params so we know if we're looking at 'all' reservations
	// or 'new' reservation, then the reservation id. We say exploded[4] coz that's the id's position
	// in the URL when we split it by slashes (admin/, reservations/, all/new/, the ID)
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	// get reservation from DB
	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminShowPostReservation updates a reservation in the admin dashboard
func (m *Repository) AdminShowPostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	// get reservation from DB
	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//data := make(map[string]interface{})
	//data["reservation"] = res
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified in the URL (if it was, it will be like this ?'y=2024&m=06')
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		//therefore the year & month are specified
		// NOTES: dates/times from the browser have to be converted for use in Go backend.
		//	Here's how (hint: use strconv.Atoi())
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))

		// NOTES: Dates in Go is handled by time.Date() wh takes a lot of params:
		//		-the converted year (converted using strconv.Atoi() as above)
		//		-the converted month (converted using strconv.Atoi() as above) which has to be converted again using time.Month()
		//		-the day in the month (1 stands for the first day-which all months have)
		//		-the hours (0)
		//		-the minutes (0)
		//		-the seconds (0)
		//		-the nano seconds (0)
		//		-the location timezone (eg time.UTC)
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	// NOTES: Prepare to pass a formatted date to the template
	data := make(map[string]interface{})
	data["now"] = now

	// NOTES: Dates in go. Here is how to add a date to a date eg here we add one month to today's date
	//		we use now.AddDate(numberOfyears, numberOfMonths, numberOfdays). A minus digit will make it in the past
	next := now.AddDate(0, 1, 0)  // next month
	last := now.AddDate(0, -1, 0) // last month

	// NOTES: How to format date/time. You can separate the bits, eg here we format month to a 2-digit
	nextMonth := next.Format("01")

	// format a year as a 4-digit year
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// NOTES: Dates in go-see techniques to get various data about dates below
	// We need to know how many days there are in each month
	// & also to get the first & last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	// We need to find a way to pass to the template data about all reserved rooms & all blocked (non-available) rooms
	for _, x := range rooms {
		// create maps to hold this data
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-02")] = 0
			blockMap[d.Format("2006-01-02")] = 0
		}

		// get all the restrictions (existing bookings) for this room, for the current month
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		// loop through the restrictions & determine whether its a reservation or a block
		//	if its a reservation, we'll put it in our reservation map, if its a block, we'll
		//	put it in our bloack map
		for _, y := range restrictions {
			if y.ReservationID > 0 {
				// its a reservation
				// reservations can be 1 day long, or 99 days long
				// we now need to loop again thru each of the dates & enter them into our reservation map
				// NOTES: In this for loop, we start from an index which will be the reservation start date,
				//	when we get to the end date (d.After(y.EndDate) == false), then we add one day
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					// put each reservation ID against its date in the reservationMap
					reservationMap[d.Format("2006-01-02")] = y.ReservationID
				}
			} else {
				// its a block
				blockMap[y.StartDate.Format("2006-01-02")] = y.RestrictionID
			}
		}

		// pass this restriction/block data to the template where it will be read & used
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		// store the blockmap for this room in the session
		// This is coz when the calendar is rendered to the screen, as the user makes changes to dates,
		//	as we go ahead to process the users's wishes, we need to know what was currently blocked before
		//	their changes, so we know how to perform the update.
		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)

	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = m.DB.UpdateProcessed(id, 1)
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminDeleteReservation deletes a reservation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = m.DB.DeleteReservation(id)
	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}
