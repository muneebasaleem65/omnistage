package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/muneebasaleem65/omnistage/internal/storage"
	"github.com/muneebasaleem65/omnistage/internal/types"
)

type Postgres struct {
	db *sql.DB
}

var _ storage.Storage = (*Postgres)(nil)

func New(storagePath string) (*Postgres, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) CreateUser(email, passwordHash, name string) (int64, error) {
	const op = "storage.postgres.CreateUser"

	var id int64

	err := p.db.QueryRow(
		`INSERT INTO USERS (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id`,
		email, passwordHash, name,
	).Scan(&id)

	//check for duplicate email error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *Postgres) GetUserByID(id int64) (types.User, error) {
	const op = "storage.postgres.GetUserByID"
	var user types.User

	err := p.db.QueryRow(
		`SELECT id, email, name, role
		FROM users
		WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return types.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}
	if err != nil {
		return types.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (p *Postgres) GetUserByEmail(email string) (types.User, error) {
	const op = "storage.postgres.GetUserByEmail"

	var user types.User

	err := p.db.QueryRow(
		`SELECT id, email, password_hash,name, role
		FROM users
		WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role)

	if errors.Is(err, sql.ErrNoRows) {
		return types.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}
	if err != nil {
		return types.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (p *Postgres) CreateVenue(name, city string, capacity int) (int64, error) {
	const op = "storage.postgres.CreateVenue"

	var id int64

	err := p.db.QueryRow(
		`INSERT INTO VENUES (name, city, capacity)
		VALUES ($1,$2,$3)
		RETURNING id`,
		name, city, capacity,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *Postgres) ListVenues() ([]types.Venue, error) {
	const op = "storage.postgres.ListVenues"

	rows, err := p.db.Query(
		`SELECT id, name, city, capacity, created_at FROM venues ORDER BY name`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//defer the closing of rows to ensure it is closed after we are done with it
	defer rows.Close()

	venues := []types.Venue{}

	for rows.Next() {
		var v types.Venue
		if err := rows.Scan(&v.ID, &v.Name, &v.City, &v.Capacity, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		venues = append(venues, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return venues, nil
}

func (p *Postgres) CreateEvent(venueID int64, title, artist, description string, startsAt time.Time) (int64, error) {
	const op = "storage.postgres.CreateEvent"

	var id int64

	err := p.db.QueryRow(
		//these are the column name from db
		`INSERT INTO EVENTS (venue_id, title, artist,description, starts_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id`,
		venueID, title, artist, description, startsAt,
	).Scan(&id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrVenueNotFound)
	}

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil

}

func (p *Postgres) ListEvents(city string, from, to *time.Time, limit, offset int) ([]types.Event, int, error) {
	const op = "storage.postgres.ListEvents"
	fromClause := ` FROM events e JOIN venues v ON v.id = e.venue_id`
	where := ` WHERE e.status = 'published'`

	args := []any{} // the VALUES
	argN := 1 // the next placeholder NUMBER
	// In laravel: 
	//  $query = Event::where('status', 'published');
	// if ($city) {
	//     $query->where('city', $city);
	// }
	//Laravels ? placeholders are positional and the builder counts them for you. Postgres uses numbered placeholders — $1, $2, $3 — so you have to do the counting yourself. argN is the counter.
	if city != "" {
		where += fmt.Sprintf(" AND v.city =$%d", argN) // 1. add the condition, using the current number
		args = append(args, city) // 2. add the value at that position
		argN++ // 3. advance the counter
	}
	if from != nil {
		where += fmt.Sprintf(" AND e.starts_at >= $%d", argN)
		args = append(args, *from)
		argN++
	}
	if to != nil {
		where += fmt.Sprintf(" AND e.starts_at <= $%d", argN)
		args = append(args, *to)
		argN++
	}

	var total int
	if err := p.db.QueryRow(
		`SELECT COUNT(*)`+ fromClause+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	rowsQuery := `SELECT e.id, e.venue_id, e.title, e.artist, e.description,
	                     e.starts_at, e.status, e.created_at` +
		fromClause + where + fmt.Sprintf(" ORDER BY e.starts_at LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, limit, offset)

	rows, err := p.db.Query(rowsQuery, args...)

	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	events := []types.Event{}
	for rows.Next() {
		var e types.Event
		if err := rows.Scan(&e.ID, &e.VenueID, &e.Title, &e.Artist, &e.Description, &e.StartsAt, &e.Status, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("%s : %w", op, err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return events, total, nil
}

func (p *Postgres) GetEventByID(id int64) (types.Event, error) {
	const op = "storage.postgres.GetEventByID"
	var event types.Event

	err := p.db.QueryRow(
		`SELECT id, venue_id, title, artist, description, starts_at, status, created_at
		FROM events
		WHERE id = $1`,
		id,
	).Scan(&event.ID, &event.VenueID, &event.Title, &event.Artist, &event.Description, &event.StartsAt, &event.Status, &event.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Event{}, fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
	}
	if err != nil {
		return types.Event{}, fmt.Errorf("%s: %w", op, err)
	}
	return event, nil
}

func (p *Postgres) UpdateEvent(id int64, title, artist, description, status *string, startsAt *time.Time) error {
	const op = "storage.postgres.UpdateEvent"

	setClauses := []string{}
	args := []any{} // the VALUES
	argN := 1 // the next placeholder NUMBER

	if title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argN))
		args = append(args, *title)
		argN++
	}

	if artist != nil {
		setClauses = append(setClauses, fmt.Sprintf("artist = $%d", argN))
		args = append(args, *artist)
		argN++	
	}

	if description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argN))
		args = append(args, *description)
		argN++
	}

	if status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argN))
		args = append(args, *status)
		argN++
	}

	if startsAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("starts_at = $%d", argN))
		args = append(args, *startsAt)
		argN++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrNoFieldsToUpdate)
	}

	setClauses = append(setClauses, "updated_at = NOW()") // Update the updated_at timestamp

	query := fmt.Sprintf(
		"UPDATE events SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		argN,
	)
	args = append(args, id)
	
	res, err := p.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
	}
	return nil
}