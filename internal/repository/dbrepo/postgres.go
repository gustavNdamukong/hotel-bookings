package dbrepo

import (
	"context"
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation to the DB
// NOTES: to return multiple values, comma-separate them in parentheses eg (int, error) below.
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	/*
	 NOTES: Once the DB connectionn is open, we want to be able to close it when its done doing its job, or if it crashes,  or times out.
	 To do so, Go has a concept of 'context' which you set with a timeout for it to be cancelled. Lets go for a 3 seconds timeout.
	 Further below where we make the DB execution, instead of using 'DB.Exec()' like so: '_, err := m.DB.Exec(stmt, ...)' which knows
	 nothing about context, use 'DB.ExecContext(ctx, stmt, ...)' or even 'DB.QueryRowContext(ctx, stmt, ...)'.

	 Only with 'DB.QueryRowContext()' can you chain a Scan() method after to extract a returned insert ID into a variable for later use.
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	// NOTES: This is how to get the last inserted record ID in postgreSQL
	stmt := `INSERT INTO reservations (first_name, last_name, email, phone, start_date, 
			end_date, room_id, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	// we return 0 for no last inserted ID returned
	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the DB
func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// NOTES: This is how to get the last inserted record ID in postgreSQL
	stmt := `INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id, 
			created_at, updated_at, restriction_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		res.ReservationID,
		time.Now(),
		time.Now(),
		res.RestrictionID,
	)

	// we return 0 for no last inserted ID returned
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomId returns true if availability exists for roomID & false if no availability exists
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomId(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `
		SELECT count(id) FROM room_restrictions 
		WHERE room_id = $1
		AND NOT ($2 > end_date OR $3 < start_date);`

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms if any for given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// initialise an empty slice
	var rooms []models.Room

	query := `
		SELECT r.id, r.room_name 
		FROM rooms r 
		WHERE r.id NOT IN (
			SELECT rr.room_id FROM room_restrictions rr WHERE $1 < rr.end_date AND $2 > rr.start_date
			);
		`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	// NOTES: Here's how u loop thru multiple values returned from DB & retrieve their values
	for rows.Next() {
		//initialise an empty room model to fill with your retrieved rooms & return
		var room models.Room
		// scan the result set into your (room) model
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}

		// add the room model to the rooms slice
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

// GetRoomById returns a room by ID
func (m *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `
		SELECT id, room_name, created_at, updated_at
		FROM rooms
		WHERE id = $1;
		`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.Created_at,
		&room.Updated_at,
	)

	if err != nil {
		return room, err
	}

	return room, nil
}
