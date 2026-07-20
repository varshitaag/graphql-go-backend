# Notes App — Backend (Go + gqlgen + Postgres)

## One important note before running

This project ships everything **except** `graph/generated.go` — that file is
the gqlgen execution engine, machine-generated from `schema.graphqls` and
usually thousands of lines long. You need to generate it locally once:

```bash
go mod tidy
go run github.com/99designs/gqlgen generate
```

This reads `gqlgen.yml` + `schema.graphqls` and produces `graph/generated.go`
(plus refreshes `graph/model/models_gen.go`). Re-run it any time you change
the schema.

## Running locally

```bash
# 1. Start Postgres (or use your own instance)
docker compose up -d postgres

# 2. Run the migration
psql "postgres://notes:notes@localhost:5432/notesdb?sslmode=disable" \
  -f internal/database/migrations/000001_create_notes_table.up.sql

# 3. Set env vars
cp .env.example .env

# 4. Generate gqlgen code (see above), then run the server
go run ./cmd/server
```

Server starts at `http://localhost:8080` with a GraphQL Playground at `/`
and the actual endpoint at `/query`.

## Example queries

```graphql
mutation {
  createNote(input: { title: "Groceries", content: "Milk, eggs, bread" }) {
    id
    title
  }
}

query {
  notes {
    id
    title
    content
    updatedAt
  }
}
```
# notes-graphql

# Adding "Sign in with Google" — File-by-File Guide

Diff (against your existing JWT-auth version): `notes-app-google-login.diff`
Full updated repo: `notes-graphql-google-login.zip`

## How this actually works, in plain terms

You do **not** implement Google's OAuth redirect flow on your server. Instead:

1. Your **frontend** loads Google's small JS library and renders a
   "Sign in with Google" button.
2. User clicks it → Google shows the account picker (the popup you're
   picturing) → user chooses an account.
3. Google hands your frontend back a signed **ID token** — a JWT that says
   "this is definitely `asha@gmail.com`, verified by Google."
4. Your frontend sends that token to a new mutation: `loginWithGoogle(idToken: ...)`.
5. Your **backend** verifies the token really came from Google and wasn't
   tampered with, pulls the email/name out of it, and either logs the user
   in, links their Google account to an existing password account, or
   creates a brand-new account — then issues your own app JWT exactly like
   `login`/`register` already do.

Your server never sees the user's Google password, never talks to Google
during login except to fetch Google's public keys (to check the
signature) — it's purely verifying a token, not orchestrating a login flow.

---

## One-time setup (do this before touching code)

1. Go to [Google Cloud Console](https://console.cloud.google.com/) →
   **APIs & Services → Credentials**.
2. Create an **OAuth 2.0 Client ID**, type "Web application."
3. Add your frontend's URL (e.g. `http://localhost:3000`) under
   **Authorized JavaScript origins**.
4. Copy the generated **Client ID** — looks like
   `123456-abc.apps.googleusercontent.com`. This goes in `GOOGLE_CLIENT_ID`
   on both frontend and backend.

---

## Backend changes

### New: `internal/database/migrations/000004_add_google_id_to_users.up.sql` / `.down.sql`
Adds a nullable, unique `google_id` column to `users`, and makes
`password_hash` nullable — a user who only ever signs in via Google never
sets a password, so it can't stay `NOT NULL`.

### `graph/schema.graphqls`
Added one line to `Mutation`:
```graphql
loginWithGoogle(idToken: String!): AuthPayload!
```
No new input type needed — it's a single scalar argument, unlike
`register`/`login` which bundle multiple fields.

### New: `internal/auth/google.go`
`VerifyGoogleToken(ctx, idToken, clientID)` — uses Google's official
`idtoken` package, which fetches Google's public signing keys, checks the
token's signature/expiry/issuer, and confirms the `audience` claim matches
your `clientID` (this is what stops someone handing you a token meant for
a *different* app). Returns the user's Google ID, email, name, and whether
Google itself has verified that email.

### `internal/repository/user_repo.go` — three new methods
- `FindByGoogleID` — is this Google account already linked to a user?
- `CreateWithGoogle` — make a new, password-less user.
- `LinkGoogleID` — attach a Google account to a user who signed up the
  normal way first (so they can use either method afterward).

### `internal/service/auth_service.go`
- `AuthService` now also holds `googleClientID`.
- New `LoginWithGoogle` method — the actual decision logic, handling three
  cases in order: already linked → issue token; email matches an existing
  password account → link it, then issue token; brand new → create account.
- Rejects the login if `EmailVerified` is `false` on the Google token —
  Google technically allows unverified emails in some edge cases, and you
  don't want to trust those for account matching.

### `graph/schema.resolvers.go`
One new resolver, mirroring `Login`:
```go
func (r *mutationResolver) LoginWithGoogle(ctx context.Context, idToken string) (*model.AuthPayload, error) {
    return r.AuthService.LoginWithGoogle(ctx, idToken)
}
```

### `internal/config/config.go` / `cmd/server/main.go`
Added `GoogleClientID`, loaded from `GOOGLE_CLIENT_ID`, required at
startup — and passed into `NewAuthService`.

### `go.mod`
Added `google.golang.org/api` (for the `idtoken` package). Run
`go mod tidy` after pulling.

### `.env.example`
Added `GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com`.

---

## Frontend changes

You'll need Google's Identity Services script and a small callback. Add
this to your `index.html`:

```html
<script src="https://accounts.google.com/gsi/client" async defer></script>

<div id="g_id_onload"
     data-client_id="YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com"
     data-callback="handleGoogleLogin">
</div>
<div class="g_id_signin" data-type="standard"></div>
```

And in `app.js`, the callback Google invokes once the user picks an account:

```javascript
async function handleGoogleLogin(response) {
  // response.credential is the ID token — send it straight to your API
  const data = await gql(`
    mutation($idToken: String!) {
      loginWithGoogle(idToken: $idToken) {
        token
        user { id name email }
      }
    }
  `, { idToken: response.credential });

  localStorage.setItem("authToken", data.loginWithGoogle.token);
  // ...then attach that token as an Authorization header on future requests
}

// Google calls this by name (data-callback="handleGoogleLogin" above),
// so it needs to be reachable globally:
window.handleGoogleLogin = handleGoogleLogin;
```

That's the whole frontend piece — no redirect handling, no popup-management
code, Google's library does that part.

---

## Testing it

Since GraphQL Playground can't do the Google popup for you, you need a real
ID token first. The easiest way: drop the HTML snippet above into a plain
page, sign in, and `console.log(response.credential)` in the callback to
grab the token. Then in Playground:

```graphql
mutation {
  loginWithGoogle(idToken: "PASTE_LONG_TOKEN_HERE") {
    token
    user {
      id
      name
      email
    }
  }
}
```

Run it twice with the same Google account — second time should hit the
"already linked" case and return the same `user.id`, not create a duplicate.

Also worth testing: register normally with `asha@example.com` via `register`,
then log in with Google using an account with that same email — should link
to the *existing* user rather than creating a second one (check `user.id`
matches).