package render

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"html/template"

	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/justinas/nosurf"
)

// This will hold all custom functions that we would want to create and make
// available to our golang templates
var functions = template.FuncMap{}

var app *config.AppConfig

// When we run the app in the browser its fine, but when running eg tests, tests are run
// from a different dir relatively, so we do this to ensure template files are always accessed
// as an absolute path
var pathToTemplates = "./templates"

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

// AddDefaultData will be used to pass to views data that should be sent to all views by default
// PopString is a built-in method on the Session library which puts something in the session
// which only lasts until the page is refreshed
func AddDefaultData(tData *models.TemplateData, request *http.Request) *models.TemplateData {
	tData.Flash = app.Session.PopString(request.Context(), "flash")
	tData.Error = app.Session.PopString(request.Context(), "error")
	tData.Warning = app.Session.PopString(request.Context(), "warning")

	/////tData.StringMap["defaultAppTitle"] = app.DefaultAppTitle
	tData.CSRFToken = nosurf.Token(request) //this will be used by all views with forms
	return tData
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, request *http.Request, requestedTemplateName string, tData *models.TemplateData) error {

	var templateCache map[string]*template.Template
	//if in development env
	if app.UseCache {
		// get the template cache from the app config instead of CreateTemplateCache() (which parses all templates anew)
		templateCache = app.TemplateCache
	} else {
		// this is just used for testing, so that we rebuild the cache on every request
		templateCache, _ = CreateTemplateCache()
	}

	//get the requested template from cache
	parsedTemplate, ok := templateCache[requestedTemplateName]
	if !ok {
		//cannot get template from cache
		//log.Fatal("Could not get template from template cache")
		return errors.New("Cannot get template from template cache")
	}

	buffer := new(bytes.Buffer)

	// beside your custom data for views, allow any other default data that should
	// be passed to the views to be added to the template data destined for the view
	tData = AddDefaultData(tData, request)

	//we do not have do go via the buffer, but we do it for fine-grained
	//control over being able to tell where a potential error may be coming from
	err := parsedTemplate.Execute(buffer, tData)
	if err != nil {
		log.Fatal(err)
	}

	//render the template
	_, err = buffer.WriteTo(w)

	if err != nil {
		fmt.Println("error writing template to browser", err)
		return err
	}

	/*
		With this approach, you do no longer need to keep track of how many files are in
		your templates directory, or how many are using a particular extension like
		a .page, .tmpl or .layout etc. All that will happen automatically
	*/
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	//Note that this function should return a pointer to *template.Template so:
	myCache := map[string]*template.Template{}

	//This function should cache all your chacheable assets in one place
	//It's recommended to first parse template files before their associated layout files
	//get all files named *.page.tmpl from the ./templates directory
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		//return whatever the current value of myCache is
		return myCache, err
	}

	//range through all files ending with *.page.tmpl
	for _, page := range pages {
		//extract just the actual file name from the full path (since pages come as the full paths to the files)
		name := filepath.Base(page)
		//parse the file & store it in a template called 'name'
		templateSet, err := template.New(name).ParseFiles(page)
		if err != nil {
			//return whatever the current value of myCache is
			return myCache, err
		}

		//parse the layout files too
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			//return whatever the current value of myCache is
			return myCache, err
		}

		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				//return whatever the current value of myCache is
				return myCache, err
			}
		}

		//add the final resulting template to our map
		myCache[name] = templateSet
	}

	return myCache, nil
}
