import { create } from 'zustand';
import type { User } from '../types';
import { api } from '../lib/api';

interface AuthState {
  user: User | null;
  token: string | null;
  loading: boolean;
  setToken: (token: string) => void;
  fetchUser: () => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  loading: true,

  setToken: (token: string) => {
    localStorage.setItem('token', token);
    set({ token });
  },

  fetchUser: async () => {
    try {
      const user = await api.getMe();
      set({ user, loading: false });
    } catch {
      localStorage.removeItem('token');
      set({ user: null, token: null, loading: false });
    }
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null });
  },
}));
