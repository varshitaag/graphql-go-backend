package auth

import (
	"context"
	"errors"

	"google.golang.org/api/idtoken"
)

var ErrInvalidGoogleToken = errors.New("invalid google id token")

type GoogleClaims struct {
	GoogleID      string
	Email         string
	EmailVerified bool
	Name          string
}

// VerifyGoogleToken validates a Google-issued ID token — the token the
// frontend receives after the user picks their Google account — and
// extracts the claims we need.
//
// idtoken.Validate does the heavy lifting: it fetches Google's public keys,
// checks the token's signature, expiry, and issuer, and confirms the
// "audience" claim matches clientID (your app's OAuth Client ID). That last
// check is what stops someone handing your server a token that was actually
// issued for a *different* app.
func VerifyGoogleToken(ctx context.Context, idTokenString, clientID string) (*GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, idTokenString, clientID)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)

	return &GoogleClaims{
		GoogleID:      payload.Subject, // Google's stable, unique ID for this user
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
	}, nil
}
