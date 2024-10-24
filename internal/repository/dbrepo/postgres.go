package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at
		FROM users WHERE id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.Created_at,
		&u.Updated_at,
	)

	if err != nil {
		return u, err
	}
	return u, nil

}

// UpdateUser a user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE users SET first_name = $1, 
		last_name = $2, 
		email = $3, 
		access_level = $4, 
		updated_at = $5
		FROM users WHERE id = $1`
	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}
	return nil
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := `SELECT id, password
		FROM users WHERE email = $1`
	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(&id, &hashedPassword)

	if err != nil {
		return id, "", err
	}

	// now compare their password with password in the system
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password!")
	} else if err != nil {
		return 0, "", err
	}
	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		FROM reservations r 
		LEFT JOIN rooms rm 
		ON (r.room_id = rm.id)
		ORDER BY r.start_date ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomId,
			&i.Created_at,
			&i.Updated_at,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// AllNewReservations returns a slice of all new reservations
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM reservations r 
		LEFT JOIN rooms rm 
		ON (r.room_id = rm.id)
		WHERE processed = 0
		ORDER BY r.start_date ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomId,
			&i.Created_at,
			&i.Updated_at,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// GetReservationById gets one reservation by its ID
func (m *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm
		ON (r.room_id = rm.id) 
		WHERE r.id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomId,
		&res.Created_at,
		&res.Updated_at,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}
	return res, nil

}

// UpdateReservation updates a reservation in the database
func (m *postgresDBRepo) UpdateReservation(res models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE reservations SET 
		first_name = $1, 
		last_name = $2, 
		email = $3, 
		phone = $4,  
		updated_at = $5
		WHERE id = $6`
	_, err := m.DB.ExecContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
	)

	if err != nil {
		return err
	}
	return nil
}

// DeleteReservation deletes a reservation by id
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM reservations 
		WHERE id = $1`
	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}
	return nil
}

// UpdateProcessed updates processed field for a reservation by id
func (m *postgresDBRepo) UpdateProcessed(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE reservations SET 
		processed = $1 
		WHERE id = $2`
	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}
	return nil
}

// AllRooms returns all rooms
func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		SELECT id, room_name, created_at, updated_at
		FROM rooms
		ORDER BY room_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.Created_at,
			&rm.Updated_at,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	// NOTES: DB here is how we use coalesce. In the 'room_restrictions' table reservation_id
	//	will not always be present & will cause errors when selecting & scanning its value in go.
	//	Hence we use coalesce to ensure if the value of that field is NULL, we default it to 0.

	/* MODIFIED THIS TO QUERY BELOW - NEEDS TESTING
	query := `
		SELECT id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		FROM room_restrictions
		WHERE $1 < end_date
		AND $2 >= start_date
		AND room_id = $3`
	*/

	query := `
	SELECT id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
	FROM room_restrictions
	WHERE NOT ($1 > end_date OR $2 < start_date)
	AND room_id = $3`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rr models.RoomRestriction
		err := rows.Scan(
			&rr.ID,
			&rr.ReservationID,
			&rr.RestrictionID,
			&rr.RoomId,
			&rr.StartDate,
			&rr.EndDate,
		)
		if err != nil {
			return nil, err
		}
		restrictions = append(restrictions, rr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}
