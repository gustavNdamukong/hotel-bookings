package dbrepo

import (
	"errors"
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation to the DB
// NOTES: to return multiple values from a func, comma-separate them in parentheses eg (int, error) below.
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room id is 2, then fail; otherwise, pass
	if res.RoomId == 2 {
		return 0, errors.New("Some error")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the DB
func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	if res.RoomId == 1000 {
		return errors.New("Some error")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomId returns true if availability exists for roomID & false if no availability exists
func (m *testDBRepo) SearchAvailabilityByDatesByRoomId(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms if any for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil

}

// GetRoomById returns a room by ID
func (m *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room

	if id > 2 {
		return room, errors.New("Some error")
	}

	return room, nil
}

func (m *testDBRepo) GetUserById(id int) (models.User, error) {
	var u models.User

	return u, nil
}

func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "me@here.ca" {
		return 1, "", nil
	}
	return 0, "", errors.New("some error")
}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) GetReservationById(id int) (models.Reservation, error) {
	var res models.Reservation
	return res, nil
}

func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (m *testDBRepo) UpdateProcessed(id, processed int) error {
	return nil
}

// AllRooms returns all rooms
func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

func (m *testDBRepo) GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction
	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

// DeleteBlockById deletes a room restriction
func (m *testDBRepo) DeleteBlockById(id int) error {
	return nil
}
