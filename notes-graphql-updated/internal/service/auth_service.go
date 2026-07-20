package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"notes-app/graph/model"
	"notes-app/internal/auth"
	"notes-app/internal/repository"
	apperrors "notes-app/pkg/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const tokenTTL = 24 * time.Hour

type AuthService struct {
	userRepo       *repository.UserRepository
	jwtSecret      string
	googleClientID string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret, googleClientID string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret, googleClientID: googleClientID}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.AuthPayload, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	if name == "" {
		return nil, apperrors.Invalid("name cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return nil, apperrors.Invalid("invalid email address")
	}
	if len(password) < 8 {
		return nil, apperrors.Invalid("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, name, email, string(hash))
	if err != nil {
		// already a *apperrors.AppError (Conflict) if it was a unique violation
		return nil, err
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &model.AuthPayload{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, passwordHash, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Deliberately generic — don't reveal whether the email exists.
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &model.AuthPayload{Token: token, User: user}, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NotFound("user not found")
	}
	return user, nil
}

// LoginWithGoogle handles three cases:
//  1. This Google account has logged in before -> just issue a token.
//  2. No Google link yet, but the email matches an existing password
//     account -> link the Google account to it, then issue a token.
//  3. Brand new email -> create a fresh, password-less account.
func (s *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (*model.AuthPayload, error) {
	claims, err := auth.VerifyGoogleToken(ctx, idToken, s.googleClientID)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid google token")
	}
	if !claims.EmailVerified {
		return nil, apperrors.Unauthorized("google email is not verified")
	}

	email := strings.ToLower(claims.Email)

	// Case 1: already linked.
	if user, err := s.userRepo.FindByGoogleID(ctx, claims.GoogleID); err == nil {
		return s.issueToken(user)
	}

	// Case 2: existing password account with the same email -> link it.
	if existing, _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		if err := s.userRepo.LinkGoogleID(ctx, existing.ID, claims.GoogleID); err != nil {
			return nil, err
		}
		return s.issueToken(existing)
	}

	// Case 3: brand new user.
	name := claims.Name
	if name == "" {
		name = email
	}
	newUser, err := s.userRepo.CreateWithGoogle(ctx, name, email, claims.GoogleID)
	if err != nil {
		return nil, err
	}
	return s.issueToken(newUser)
}

func (s *AuthService) issueToken(user *model.User) (*model.AuthPayload, error) {
	token, err := auth.GenerateToken(user.ID, s.jwtSecret, tokenTTL)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{Token: token, User: user}, nil
}
