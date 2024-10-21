package models

import (
	"time"
)

// These models will be useful to us once we start writing our DB methods.
// Reservation holds reservation data
// DOC: make the properties start in uppercase
// so they can be accessible outside of this package

// DOC: It is also possible to pass a model into another model here as a property
// if you want to.
// By convention, model names should be in singular while the corresponding DB
//	tables should be in plural

// User is the user model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	Created_at  time.Time
	Updated_at  time.Time
}

// Room is the room model
type Room struct {
	ID         int
	RoomName   string
	Created_at time.Time
	Updated_at time.Time
}

// Room is the room model
type Restriction struct {
	ID              int
	RestrictionName string
	Created_at      time.Time
	Updated_at      time.Time
}

// Reservation is the Reservation model
type Reservation struct {
	ID         int
	FirstName  string
	LastName   string
	Email      string
	Phone      string
	StartDate  time.Time
	EndDate    time.Time
	RoomId     int
	Created_at time.Time
	Updated_at time.Time
	Room       Room
	Processed  int
}

// RoomRestriction is the RoomRestriction model
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomId        int
	ReservationID int
	RestrictionID int
	Created_at    time.Time
	Updated_at    time.Time
	Room          Room
	// DOC: we may not need these, but we place them here in case we need them
	Reservation Reservation
	Restriction Restriction
}

// MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
