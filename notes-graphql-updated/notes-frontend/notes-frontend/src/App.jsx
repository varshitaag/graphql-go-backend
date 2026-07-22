import { AuthProvider, useAuth } from "./context/AuthContext";
import LoginPage from "./pages/LoginPage";
import NotesPage from "./pages/NotesPage";

// This is the whole "routing" logic for a 2-page app: if there's a token,
// show notes; if not, show login. No react-router needed at this size —
// adding one would be one more thing to explain for no real benefit here.
function AppContent() {
  const { token } = useAuth();
  return token ? <NotesPage /> : <LoginPage />;
}

export default function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}
