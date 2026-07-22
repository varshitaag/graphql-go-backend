# Notes App — React Frontend

## Folder structure

```
notes-frontend/
├── package.json           — dependencies (just React + Vite, nothing else)
├── vite.config.js          — dev server / build config
├── index.html              — the single HTML page React mounts into
├── .env.example             — backend URL + Google Client ID
└── src/
    ├── main.jsx             — entry point, renders <App /> into the page
    ├── App.jsx              — decides: show LoginPage or NotesPage?
    ├── index.css            — all styling, both pages
    ├── api/
    │   └── graphql.js       — ⭐ THE connection to your backend — every network call lives here
    ├── context/
    │   └── AuthContext.jsx  — holds the logged-in user + token in memory + localStorage
    └── pages/
        ├── LoginPage.jsx    — email/password + Register toggle + Google button
        └── NotesPage.jsx    — list/create/edit/delete notes
```

No React Router. With only two "pages," `App.jsx` just checks
"is there a token?" and renders one component or the other — adding a
router would be one more concept to learn for no real benefit at this size.

---

## Where this connects to your Go backend

**Everything goes through one file: `src/api/graphql.js`.** No component
ever calls `fetch()` directly — they all import functions like
`loginUser()`, `fetchNotes()`, `createNote()` from that file. If you ever
change your backend's URL, or add a new field to a query, this is the only
file you touch.

```js
const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/query";
```
This is the literal URL of your GraphQL endpoint — the same `/query` path
your Go server mounts in `cmd/server/main.go`.

Every function in that file sends a `POST` request shaped like:
```json
{ "query": "mutation { ... }", "variables": { ... } }
```
which is exactly what your `srv := handler.NewDefaultServer(...)` in
`main.go` expects to receive.

**Auth flows through the same file too** — look at `gqlRequest`:
```js
if (token) {
  headers["Authorization"] = `Bearer ${token}`;
}
```
This header is what your backend's `internal/auth/middleware.go` reads on
every request, to figure out which user is calling. Compare the two sides:

| Frontend | Backend |
|---|---|
| `AuthContext` stores `token` after login | `auth.Middleware` reads the `Authorization` header |
| `fetchNotes(token)` attaches it | `auth.UserIDFromContext(ctx)` extracts the user ID from it |
| Notes returned belong to that user | `note_repo.go` filters every query by `WHERE user_id = ...` |

So the "a user can only see their own notes" guarantee you built on the
backend works automatically here — the frontend never sends a user ID
itself, it just sends whichever token is currently stored, and the backend
decides who that is.

---

## How each requested feature maps to code

**"Email and password for logging in"** → `LoginPage.jsx`, the `<form>`
with `email`/`password` inputs, calling `loginUser()` from `graphql.js` on
submit.

**"Button for registering"** → same page, `mode` state toggles between
`"login"` and `"register"`. In register mode, an extra `name` field
appears and the form calls `registerUser()` instead.

**"Button for Google login"** → the `<div ref={googleButtonRef}>` at the
bottom of `LoginPage.jsx`. Google's own script renders an actual button
inside it (you don't build that button yourself) — see the `useEffect` at
the top of the file for how it's loaded and initialized.

**"If Google login is clicked, without password it should go to the next
page"** → `handleGoogleResponse()` in `LoginPage.jsx`. Google calls this
function directly once the user picks an account, handing it a signed
token — no password is ever asked for. That function calls
`loginWithGoogle()`, then `login(token, user)` from `AuthContext`, which
sets `token` in state — and since `App.jsx` renders `NotesPage` whenever
`token` is truthy, the page switch happens automatically, with zero
routing code needed.

**"Notes page: create, list, update, delete, his own notes alone"** →
`NotesPage.jsx`. `loadNotes()` runs on mount and calls `fetchNotes(token)`;
create/edit share one form (`editingId` being `null` or not decides which);
delete calls `deleteNote(token, id)` then reloads the list. The "his own
notes alone" part isn't anything this page does explicitly — it's a side
effect of always sending `token`, and the backend enforcing the rest.

---

## Setup

### 1. Backend: add CORS (required — see note below)

Your backend has no CORS headers configured. Since the React dev server
runs on a different port than your Go server, the browser will block every
request until you add this. A new file `cors.go` is included alongside
this doc — put it in `internal/auth/` next to your existing `middleware.go`.

Then in `cmd/server/main.go`, wrap the handler one more time:

```go
handlerWithAuth := auth.Middleware(cfg.JWTSecret)(mux)
handlerWithCORS := auth.CORS("http://localhost:5173")(handlerWithAuth)

log.Printf("🚀 server ready at http://localhost:%s/", cfg.Port)
log.Fatal(http.ListenAndServe(":"+cfg.Port, handlerWithCORS))
```
(`5173` is Vite's default dev server port — adjust if yours differs.)

### 2. Frontend

```bash
cd notes-frontend
npm install
cp .env.example .env
```

Edit `.env`:
```
VITE_API_URL=http://localhost:8080/query
VITE_GOOGLE_CLIENT_ID=your-web-client-id.apps.googleusercontent.com
```

**Important:** this must be your **web application** OAuth Client ID (the
kind you'd have set up originally, with `http://localhost:5173` added
under "Authorized JavaScript origins" in Google Cloud Console) — not the
"TV and Limited Input devices" one used earlier for terminal-only testing.
Google's JS button only works with a web-type client.

```bash
npm run dev
```

Opens at `http://localhost:5173`. Make sure your Go backend is running
first (`go run ./cmd/server`) — refresh the page if you started the
frontend before the backend was up.

---

## Testing it end-to-end

1. Open `http://localhost:5173` — should show the login card.
2. Click "New here? Register," fill in name/email/password, submit — should
   land directly on the notes page (empty list).
3. Refresh the page — should **stay** on the notes page (this confirms
   `AuthContext` reading from `localStorage` on load is working).
4. Click "Log out" — back to login page. Log back in with the same
   email/password — should work.
5. Click the Google button, pick an account — should go straight to notes,
   no password step.
6. Create a note, edit it, delete it — confirm each round-trips through
   the list correctly.
7. Log out, register a *second* account, create a note there — confirm the
   first account's notes never appear for the second account.
