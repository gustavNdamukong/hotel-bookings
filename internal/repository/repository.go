package repository

import (
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	// Write a reservation to the DB
	// NOTES: to return multiple values, comma-separate them in parentheses eg (int, error) below.
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomId(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomById(id int) (models.Room, error)
}
