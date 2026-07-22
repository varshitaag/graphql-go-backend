// This file is the single connection point to the backend.
// Every GraphQL call in the whole app goes through gqlRequest() below —
// nothing else in the codebase touches fetch() or knows the API's URL.

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/query";

/**
 * Sends a GraphQL query/mutation to the backend's /query endpoint.
 * If a token is passed, it's attached as "Authorization: Bearer <token>" —
 * this is how the backend's JWT middleware identifies which user is making
 * the request (see internal/auth/middleware.go on the Go side).
 */
async function gqlRequest(query, variables = {}, token = null) {
  const headers = { "Content-Type": "application/json" };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(API_URL, {
    method: "POST",
    headers,
    body: JSON.stringify({ query, variables }),
  });

  const json = await res.json();

  // GraphQL always returns 200 OK even on failure — errors show up in
  // json.errors instead of the HTTP status. Always check this first.
  if (json.errors) {
    throw new Error(json.errors[0].message);
  }

  return json.data;
}

// ---- Auth ----

export function registerUser(name, email, password) {
  return gqlRequest(
    `mutation($input: RegisterInput!) {
      register(input: $input) {
        token
        user { id name email }
      }
    }`,
    { input: { name, email, password } }
  );
}

export function loginUser(email, password) {
  return gqlRequest(
    `mutation($input: LoginInput!) {
      login(input: $input) {
        token
        user { id name email }
      }
    }`,
    { input: { email, password } }
  );
}

export function loginWithGoogle(idToken) {
  return gqlRequest(
    `mutation($idToken: String!) {
      loginWithGoogle(idToken: $idToken) {
        token
        user { id name email }
      }
    }`,
    { idToken }
  );
}

// ---- Notes (all require a token — the backend rejects these without one) ----

export function fetchNotes(token) {
  return gqlRequest(
    `query {
      notes { id title content createdAt updatedAt }
    }`,
    {},
    token
  );
}

export function createNote(token, title, content) {
  return gqlRequest(
    `mutation($input: NewNote!) {
      createNote(input: $input) { id title content }
    }`,
    { input: { title, content } },
    token
  );
}

export function updateNote(token, id, title, content) {
  return gqlRequest(
    `mutation($id: ID!, $input: UpdateNoteInput!) {
      updateNote(id: $id, input: $input) { id title content }
    }`,
    { id, input: { title, content } },
    token
  );
}

export function deleteNote(token, id) {
  return gqlRequest(
    `mutation($id: ID!) {
      deleteNote(id: $id)
    }`,
    { id },
    token
  );
}
