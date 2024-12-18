package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

// Holds the application config
type AppConfig struct {
	UseCache        bool
	TemplateCache   map[string]*template.Template
	DefaultAppTitle string
	InfoLog         *log.Logger
	InProduction    bool
	Session         *scs.SessionManager
	ErrorLog        *log.Logger
	MailChan        chan models.MailData
}
