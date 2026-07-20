package service

import (
	"context"
	"strings"

	"notes-app/graph/model"
	"notes-app/internal/repository"
	apperrors "notes-app/pkg/errors"
)

type NoteService struct {
	repo *repository.NoteRepository
}

func NewNoteService(repo *repository.NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

func (s *NoteService) ListNotes(ctx context.Context, userID string) ([]*model.Note, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *NoteService) GetNote(ctx context.Context, userID, id string) (*model.Note, error) {
	note, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, apperrors.NotFound("note not found")
	}
	return note, nil
}

func (s *NoteService) CreateNote(ctx context.Context, userID, title, content string) (*model.Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, apperrors.Invalid("title cannot be empty")
	}
	return s.repo.Create(ctx, userID, title, content)
}

func (s *NoteService) UpdateNote(ctx context.Context, userID, id string, title, content *string) (*model.Note, error) {
	if title != nil && strings.TrimSpace(*title) == "" {
		return nil, apperrors.Invalid("title cannot be empty")
	}
	note, err := s.repo.Update(ctx, userID, id, title, content)
	if err != nil {
		return nil, apperrors.NotFound("note not found")
	}
	return note, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, userID, id string) (bool, error) {
	return s.repo.Delete(ctx, userID, id)
}
