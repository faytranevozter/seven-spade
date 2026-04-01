import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuthStore } from './stores/authStore';
import LoginPage from './pages/LoginPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import DashboardPage from './pages/DashboardPage';
import RoomLobbyPage from './pages/RoomLobbyPage';
import GamePage from './pages/GamePage';
import LeaderboardPage from './pages/LeaderboardPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, user, loading } = useAuthStore();

  if (!token) return <Navigate to="/login" replace />;
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-gray-400">Loading...</div>
      </div>
    );
  }
  if (!user) return <Navigate to="/login" replace />;

  return <>{children}</>;
}

export default function App() {
  const { token, fetchUser } = useAuthStore();

  useEffect(() => {
    if (token) {
      fetchUser();
    } else {
      useAuthStore.setState({ loading: false });
    }
  }, [token, fetchUser]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        <Route path="/room/:code" element={<ProtectedRoute><RoomLobbyPage /></ProtectedRoute>} />
        <Route path="/game/:code" element={<ProtectedRoute><GamePage /></ProtectedRoute>} />
        <Route path="/leaderboard" element={<ProtectedRoute><LeaderboardPage /></ProtectedRoute>} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
