import { createContext, useContext, useState } from "react";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  // Read from localStorage on first load, so refreshing the page doesn't
  // throw the user back to the login screen.
  const [token, setToken] = useState(() => localStorage.getItem("authToken"));
  const [user, setUser] = useState(() => {
    const stored = localStorage.getItem("authUser");
    return stored ? JSON.parse(stored) : null;
  });

  // Called after register / login / loginWithGoogle all succeed —
  // they all return the same { token, user } shape from the backend.
  function login(newToken, newUser) {
    localStorage.setItem("authToken", newToken);
    localStorage.setItem("authUser", JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
  }

  function logout() {
    localStorage.removeItem("authToken");
    localStorage.removeItem("authUser");
    setToken(null);
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ token, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

// Convenience hook — lets any component do `const { token, user, login, logout } = useAuth()`
export function useAuth() {
  return useContext(AuthContext);
}
