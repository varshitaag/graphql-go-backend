package graph

import "notes-app/internal/service"

// Resolver is the root struct all resolver methods hang off.
// It holds the services resolvers need — nothing more.
type Resolver struct {
	NoteService *service.NoteService
	AuthService *service.AuthService
}

func NewResolver(noteService *service.NoteService, authService *service.AuthService) *Resolver {
	return &Resolver{NoteService: noteService, AuthService: authService}
}
