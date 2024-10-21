package render

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"html/template"

	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/justinas/nosurf"
)

/*
NOTES: This will hold all custom functions that we would want to create and make
available to our golang templates. Basically; this is how to create custom funcs
for golang templates-similar to how there's a way to create custom func for other
templating engines like Twig. It is very important to note that having created the
function map like this (template.FuncMap which we've stored in the variable functions
below); that is not enough. It will only work when the function map is applied to
your template when you're parsing it. In our app here we are parsing the template in
CreateTemplateCache() below in this line:

	templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)

What does the magic to apply your custom function to the template is this call to
Funcs(functions) which is chained on to template.New(name). 'name' is the target
template file by the way. The custom function of yours will then become available
in your target view template and can be used inside of it like so:

	<td>{{ myCustomFunction .StartDate }}</td>
*/
var functions = template.FuncMap{
	"humanDate": HumanDate,
}

var app *config.AppConfig

// When we run the app in the browser its fine, but when running eg tests, tests are run
// from a different dir relatively, so we do this to ensure template files are always accessed
// as an absolute path
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// NOTES: accepts a date & returns in the format 'YYYY-MM-DD'
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// NOTES: AddDefaultData will be used to pass to views data that should be sent to all views by default
// PopString is a built-in method on the Session library which puts something in the session
// which only lasts until the page is refreshed.
// NOTES: So here we also learn how to flash data to the session
func AddDefaultData(tData *models.TemplateData, request *http.Request) *models.TemplateData {
	tData.Flash = app.Session.PopString(request.Context(), "flash")
	tData.Error = app.Session.PopString(request.Context(), "error")
	tData.Warning = app.Session.PopString(request.Context(), "warning")

	/////tData.StringMap["defaultAppTitle"] = app.DefaultAppTitle
	tData.CSRFToken = nosurf.Token(request) //this will be used by all views with forms
	// NOTES: How to check if the session contains a variable
	if app.Session.Exists(request.Context(), "user_id") {
		tData.IsAuthenticated = 1
	}
	return tData
}

// Template renders templates using html/template
func Template(w http.ResponseWriter, request *http.Request, requestedTemplateName string, tData *models.TemplateData) error {

	var templateCache map[string]*template.Template
	//if in development env
	if app.UseCache {
		// get the template cache from the app config instead of CreateTemplateCache() (which parses all templates anew)
		templateCache = app.TemplateCache
	} else {
		// this is just used for testing, so that we rebuild the cache on every request
		templateCache, _ = CreateTemplateCache()
	}

	//store the parsed template (view file) in the cache if its not there already &, or, get the requested template from cache.
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

	// render the template
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
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)
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

		//add the final resulting template to our map, which is the cache.
		myCache[name] = templateSet
	}

	return myCache, nil
}
