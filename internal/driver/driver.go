package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DOC: Connecting to DB-you should have gone into the root of your app (eg cmd/web - or whatever the case may be)
// & ran this cmd to import the DB driver package: "go get github.com/jackc/pgx/v4" & then importing it here above.
// Here our chosen driver package is pgx from a guy called 'jackc'.

// DB holds the database connection pool
// this will allow us switch to different DB types in our application
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

// max num of DB connections this application can have
const maxOpenDbConn = 10

// max num of idle DB connections allowed
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates a connection pool for postgres
func ConnectSQL(dsn string) (*DB, error) {
	newDb, err := NewDatabase(dsn)
	if err != nil {
		// DOC: panic means you want the prog to halt coz if this doesn't work, nothing else should work
		panic(err)
	}
	newDb.SetMaxOpenConns(maxOpenDbConn)
	newDb.SetMaxIdleConns(maxIdleDbConn)
	newDb.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = newDb

	err = testDB(newDb)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// testDB tries to ping the DB
func testDB(newDb *sql.DB) error {
	err := newDb.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase creates a new DB for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// DOC: good way to quickly test for & handle an error, by calling a func that
	// you know will return an error in case it fails
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
