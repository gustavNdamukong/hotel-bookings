package handlers

import (
	"log"
	"net/http"

	"github.com/gustavNdamukong/hotel-bookings/pkg/config"
	"github.com/gustavNdamukong/hotel-bookings/pkg/models"
	"github.com/gustavNdamukong/hotel-bookings/pkg/render"
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

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
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
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
