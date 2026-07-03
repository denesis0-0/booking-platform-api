package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrSlotAlreadyBooked = errors.New("slot already booked")

type Booking struct {
	ID        string    `json:"id"`
	SlotID    string    `json:"slot_id"`
	UserName  string    `json:"user_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBookingParams struct {
	SlotID   string
	UserName string
}

func (p *Postgres) CreateBooking(ctx context.Context, params CreateBookingParams) (Booking, error) {
	query := `
		INSERT INTO bookings (slot_id, user_name)
		VALUES ($1::uuid, $2)
		ON CONFLICT DO NOTHING
		RETURNING id::text, slot_id::text, user_name, status, created_at
	`

	var booking Booking

	err := p.Pool.QueryRow(
		ctx,
		query,
		params.SlotID,
		params.UserName,
	).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.UserName,
		&booking.Status,
		&booking.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Booking{}, ErrSlotAlreadyBooked
		}

		return Booking{}, err
	}

	return booking, nil
}

func (p *Postgres) ListBookings(ctx context.Context) ([]Booking, error) {
	query := `
		SELECT id::text, slot_id::text, user_name, status, created_at
		FROM bookings
		ORDER BY created_at DESC
	`

	rows, err := p.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		err := rows.Scan(
			&booking.ID,
			&booking.SlotID,
			&booking.UserName,
			&booking.Status,
			&booking.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}
