package render

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"html/template"

	"github.com/gustavNdamukong/hotel-bookings/pkg/config"
	"github.com/gustavNdamukong/hotel-bookings/pkg/models"
)

// AddDefaultData will be used to pass to views data that should be sent to all views by default
func AddDefaultData(tData *models.TemplateData) *models.TemplateData {
	/////tData.StringMap["defaultAppTitle"] = app.DefaultAppTitle
	return tData
}

// create template functions
var funcMap = template.FuncMap{}

var app *config.AppConfig

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, requestedTemplateName string, tData *models.TemplateData) {

	var templateCache map[string]*template.Template
	//if in development env
	if app.UseCache {
		// get the template cache from the app config instead of CreateTemplateCache()
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	//get the requested template from cache
	parsedTemplate, ok := templateCache[requestedTemplateName]
	if !ok {
		//cannot get template from cache
		log.Fatal("Could not get template from template cache")
	}

	buffer := new(bytes.Buffer)

	// beside your custom data for views, allow any other default data that should
	// be passed to the views to be added to the template data destined for the view
	tData = AddDefaultData(tData)

	//we do not have do go via the buffer, but we do it for fine-grained
	//control over being able to tell where a potential error may be coming from
	_ = parsedTemplate.Execute(buffer, tData)

	//render the template
	_, err := buffer.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to browser", err)
	}

	/*
		With this approach, you do no longer need to keep track of how many files are in
		your templates directory, or how many are using a particular extension like
		a .page, .tmpl or .layout etc. All that will happen automatically
	*/
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	//Note that this function should return a pointer to *template.Template so:
	myCache := map[string]*template.Template{}

	//This function should cache all your chacheable assets in one place
	//It's recommended to first parse template files before their associated layout files
	//get all files named *.page.tmpl from the ./templates directory
	pages, err := filepath.Glob("./templates/*page.tmpl")
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
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			//return whatever the current value of myCache is
			return myCache, err
		}

		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob("./templates/*.layout.tmpl")
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
