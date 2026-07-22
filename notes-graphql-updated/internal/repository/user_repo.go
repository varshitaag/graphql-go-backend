package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"notes-app/graph/model"
	apperrors "notes-app/pkg/errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at`,
		name, email, passwordHash)

	user, err := scanUser(row)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// unique_violation — either the name or the email is already taken
			return nil, apperrors.Conflict("name or email already in use")
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail returns the user plus their stored password hash, needed only
// for verifying a login attempt — the hash is never exposed via GraphQL.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, string, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at, password_hash
		FROM users WHERE email = $1`, email)

	var u model.User
	var createdAt time.Time
	var passwordHash sql.NullString

	err := row.Scan(&u.ID, &u.Name, &u.Email, &createdAt, &passwordHash)
	if err != nil {
		return nil, "", err
	}
	u.CreatedAt = createdAt.Format(time.RFC3339)
	return &u, passwordHash.String, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at
		FROM users WHERE id = $1`, id)

	return scanUser(row)
}

// FindByGoogleID looks up a user previously linked to this Google account.
func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at
		FROM users WHERE google_id = $1`, googleID)

	return scanUser(row)
}

// CreateWithGoogle creates a brand-new user signing up for the first time
// via Google — no password_hash, since they'll always authenticate through
// Google.
func (r *UserRepository) CreateWithGoogle(ctx context.Context, name, email, googleID string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, google_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at`,
		name, email, googleID)

	user, err := scanUser(row)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, apperrors.Conflict("name or email already in use")
		}
		return nil, err
	}
	return user, nil
}

// LinkGoogleID attaches a Google account to a user who originally signed up
// with email/password — so they can log in either way afterward.
func (r *UserRepository) LinkGoogleID(ctx context.Context, userID, googleID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET google_id = $2 WHERE id = $1`, userID, googleID)
	return err
}

// GetPasswordDetails retrieves the password hash and google_id for a given userID.
func (r *UserRepository) GetPasswordDetails(ctx context.Context, userID string) (string, string, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT password_hash, google_id FROM users WHERE id = $1`, userID)
	var passwordHash sql.NullString
	var googleID sql.NullString
	if err := row.Scan(&passwordHash, &googleID); err != nil {
		return "", "", err
	}
	return passwordHash.String, googleID.String, nil
}

// UpdatePasswordHash updates the password hash for a given userID.
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID string, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $2 WHERE id = $1`, userID, passwordHash)
	return err
}


func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var createdAt time.Time

	if err := row.Scan(&u.ID, &u.Name, &u.Email, &createdAt); err != nil {
		return nil, err
	}
	u.CreatedAt = createdAt.Format(time.RFC3339)
	return &u, nil
}
