import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';

export default function AuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setToken, fetchUser } = useAuthStore();

  useEffect(() => {
    const token = searchParams.get('token');
    if (token) {
      setToken(token);
      fetchUser().then(() => navigate('/dashboard'));
    } else {
      navigate('/login');
    }
  }, [searchParams, setToken, fetchUser, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="animate-pulse text-gray-400">Signing in...</div>
    </div>
  );
}
