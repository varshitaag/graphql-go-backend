import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { fetchNotes, createNote, updateNote, deleteNote } from "../api/graphql";

export default function NotesPage() {
  const { token, user, logout } = useAuth();

  const [notes, setNotes] = useState([]);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [editingId, setEditingId] = useState(null); // null = creating, else editing this note's id
  const [error, setError] = useState("");

  // Load this user's notes as soon as the page mounts. `token` comes from
  // AuthContext — the backend uses it to figure out *which* user's notes
  // to return, so we never need to pass a user ID ourselves.
  useEffect(() => {
    loadNotes();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function loadNotes() {
    try {
      const data = await fetchNotes(token);
      setNotes(data.notes);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCreate(e) {
    e.preventDefault();
    if (!title.trim()) return;
    setError("");
    try {
      await createNote(token, title, content);
      resetForm();
      loadNotes();
    } catch (err) {
      setError(err.message);
    }
  }

  function startEdit(note) {
    setEditingId(note.id);
    setTitle(note.title);
    setContent(note.content);
  }

  async function handleUpdate(e) {
    e.preventDefault();
    setError("");
    try {
      await updateNote(token, editingId, title, content);
      resetForm();
      loadNotes();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDelete(id) {
    setError("");
    try {
      await deleteNote(token, id);
      loadNotes();
    } catch (err) {
      setError(err.message);
    }
  }

  function resetForm() {
    setEditingId(null);
    setTitle("");
    setContent("");
  }

  return (
    <div className="notes-page">
      <header className="notes-header">
        <h1>Notes</h1>
        <div className="user-info">
          <span>{user?.name}</span>
          <button onClick={logout} className="link-button">
            Log out
          </button>
        </div>
      </header>

      <form onSubmit={editingId ? handleUpdate : handleCreate} className="note-form">
        <input
          type="text"
          placeholder="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
        <textarea
          placeholder="Write something..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={3}
        />
        <div className="form-actions">
          <button type="submit">{editingId ? "Save Changes" : "Add Note"}</button>
          {editingId && (
            <button type="button" onClick={resetForm} className="secondary">
              Cancel
            </button>
          )}
        </div>
      </form>

      {error && <p className="error">{error}</p>}

      <ul className="notes-list">
        {notes.length === 0 && <li className="empty">No notes yet.</li>}
        {notes.map((note) => (
          <li key={note.id} className="note">
            <p className="note-title">{note.title}</p>
            <p className="note-content">{note.content}</p>
            <div className="note-actions">
              <button onClick={() => startEdit(note)}>Edit</button>
              <button onClick={() => handleDelete(note.id)}>Delete</button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
