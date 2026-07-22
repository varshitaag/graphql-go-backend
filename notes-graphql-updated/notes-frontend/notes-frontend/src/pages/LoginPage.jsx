import { useState, useEffect, useRef } from "react";
import { useAuth } from "../context/AuthContext";
import { registerUser, loginUser, loginWithGoogle } from "../api/graphql";

const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID;

export default function LoginPage() {
  const { login } = useAuth();

  // "login" or "register" — one form, toggled between two modes, rather
  // than two separate pages, since they share every field except "name".
  const [mode, setMode] = useState("login");

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const googleButtonRef = useRef(null);

  // Load Google's Identity Services script once when this page mounts,
  // then tell it to render its own button into googleButtonRef.
  // This is the piece that shows the actual Google account picker —
  // your app code never builds that UI itself.
  useEffect(() => {
    if (!GOOGLE_CLIENT_ID) {
      console.warn("VITE_GOOGLE_CLIENT_ID is not set — Google login button won't render.");
      return;
    }

    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.onload = () => {
      window.google.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID,
        callback: handleGoogleResponse, // Google calls this once the user picks an account
      });
      window.google.accounts.id.renderButton(googleButtonRef.current, {
        theme: "outline",
        size: "large",
        width: 320,
      });
    };
    document.body.appendChild(script);

    return () => document.body.removeChild(script);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Google hands us `response.credential` — a signed ID token proving
  // who the user is. We send it straight to our backend's loginWithGoogle
  // mutation; the backend does all the actual verification.
  async function handleGoogleResponse(response) {
    setError("");
    setLoading(true);
    try {
      const data = await loginWithGoogle(response.credential);
      login(data.loginWithGoogle.token, data.loginWithGoogle.user);
      // No password step here at all — login() above immediately marks
      // the user as authenticated, which is what makes App.jsx switch
      // to the notes page.
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      if (mode === "register") {
        const data = await registerUser(name, email, password);
        login(data.register.token, data.register.user);
      } else {
        const data = await loginUser(email, password);
        login(data.login.token, data.login.user);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Notes</h1>
        <p className="subtitle">
          {mode === "login" ? "Log in to your account" : "Create a new account"}
        </p>

        <form onSubmit={handleSubmit}>
          {mode === "register" && (
            <input
              type="text"
              placeholder="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          )}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          {error && <p className="error">{error}</p>}

          <button type="submit" disabled={loading}>
            {loading ? "Please wait..." : mode === "login" ? "Log In" : "Register"}
          </button>
        </form>

        <button
          type="button"
          className="link-button"
          onClick={() => setMode(mode === "login" ? "register" : "login")}
        >
          {mode === "login" ? "New here? Register" : "Already have an account? Log in"}
        </button>

        <div className="divider">
          <span>or</span>
        </div>

        {/* Google renders its own button inside this div — we don't style it ourselves */}
        <div ref={googleButtonRef} className="google-button-slot"></div>
      </div>
    </div>
  );
}
