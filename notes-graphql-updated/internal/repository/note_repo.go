package repository

import (
	"context"
	"database/sql"
	"time"

	"notes-app/graph/model"
)

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) FindAll(ctx context.Context, userID string) ([]*model.Note, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE user_id = $1
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*model.Note
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepository) FindByID(ctx context.Context, userID, id string) (*model.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes WHERE id = $1 AND user_id = $2`, id, userID)

	return scanNote(row)
}

func (r *NoteRepository) Create(ctx context.Context, userID, title, content string) (*model.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO notes (title, content, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, content, created_at, updated_at`, title, content, userID)

	return scanNote(row)
}

func (r *NoteRepository) Update(ctx context.Context, userID, id string, title, content *string) (*model.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE notes
		SET title = COALESCE($3, title),
		    content = COALESCE($4, content),
		    updated_at = $5
		WHERE id = $1 AND user_id = $2
		RETURNING id, title, content, created_at, updated_at`,
		id, userID, title, content, time.Now())

	return scanNote(row)
}

func (r *NoteRepository) Delete(ctx context.Context, userID, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

// scanner covers both *sql.Row and *sql.Rows, so one helper works for both.
type scanner interface {
	Scan(dest ...any) error
}

func scanNote(s scanner) (*model.Note, error) {
	var n model.Note
	var createdAt, updatedAt time.Time

	err := s.Scan(&n.ID, &n.Title, &n.Content, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	n.CreatedAt = createdAt.Format(time.RFC3339)
	n.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &n, nil
}
