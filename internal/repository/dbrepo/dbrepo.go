package dbrepo

import (
	"database/sql"

	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// DOC: now we have a means to swap out different DB connection types for our application.
// If instead a postgres we wanted to create another conn repo, we'll similarly add two structures in this file:
// for example for a new mysql conn:
//		-i) a mysqlDBRepo struct
//		-ii) a NewMysqlRepo(...) func

// Notice that the repository.DatabaseRepo being returned by the NewPostgresRepo(...) func here is an interface.
// In this same directory (as this file-dbrepo), we will have a file 'postgres.go' (one for each DB conn type)
//	which will contain funcs for its corresponding DB conn type struct-in this case the 'postgresDBRepo' struct.
//	The funcs will be written exactly as defined in the interface in 'repository/repository.go'
//	These funcs will be the DB conn type's (in this case the 'postgresDBRepo') receiver funcs.

// Therefore back in your handler funcs (in 'handlers/handlers.go'), in any of your handler funcs that render
// web pages eg Home(), on the repository that they all have access to, you also now have access to the DB property,
// and what's even more exiting is that on that DB property, you will now be able to call your DB repository funcs
// which are associated to (receiver funcs of) your DB conn type structs defined here above
// (& declared in 'internal/repository/dbrepo/postgres.go' in the case of postgres DB conn) as well.

// For example, now you have in internal/repository/dbrepo/postgres.go the func 'AllUsers()' as a receiver func of
// the postgresDBRepo struct, and that 'AllUsers()' func is exactly as defined in the interface 'DatabaseRepo' that
// the the postgresDBRepo struct is associated with. So when rendering a view file in a handler func eg in
// '/internal/handlers/handlers.go' in any of the funcs eg Home() you can modify it
/*
	FROM:

		func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
			render.Template(w, r, "index.page.tmpl", &models.TemplateData{})
		}

	TO:

		func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
			var users := m.DB.AllUsers()
			render.Template(w, r, "index.page.tmpl", &models.TemplateData{})
		}

*/

// Repository for testing that includes DB simulation
type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
