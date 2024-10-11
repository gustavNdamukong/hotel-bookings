package render

import (
	"net/http"
	"testing"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

// As we are trying to test the the 'AddDefaultData()' function declared in render.go
func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Fatal(err)
	}

	// we need something to test, so as an example, let's put some info into our flash
	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}

}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	templateCache, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = templateCache

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	//	NOTES: we need a response writer and a request that RenderTemplate will use to generate the view/template
	// file, hence we create a writer (myWriter in the test setup to use here) and a request above (r, err := getSession())
	//	a type response writer must meet 3 criterion before it can pass as a response writer:
	// it is an interface that has a) a header, b) a writer c) and a write header method
	// If you can create a type that fulfills all those 3 things, you can use that type in your app as
	// a response writer, and the way to accomplish that in testing is to create a struct eg 'myWriter', then
	// create a receiver function for it as we have done in 'render/setup_test.go' by creating for the struct type 'myWriter'
	// threee methods: 'Header()', 'WriteHeader(i int)' & 'Write()'
	var ww myWriter

	// TODO: this was 'RenderTemplate' before. Confirm that the change is ok
	err = Template(&ww, r, "index.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser", err)
	}

	// TODO: this was 'RenderTemplate' before. Confirm that the change is ok
	err = Template(&ww, r, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist")
	}

}

// getSession() creates and returns a session object
func getSession() (*http.Request, error) {
	// NOTES: this is how we create a request within a test environment (using 'http.NewRequest()')
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	// NOTES: we really need to create a context, & we get that from the request we just created above
	// we will then proceed to use that context every time we write or read from the session
	ctx := r.Context()
	// put session data into that context
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	// having added the session to the context, we need to then put the context back into the request
	//then return the request top the caller
	r = r.WithContext(ctx)
	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
