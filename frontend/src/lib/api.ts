import type { User, Room, LeaderboardEntry } from '../types';

const BASE = '';

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token');
  const res = await fetch(`${BASE}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  getMe: () => request<User>('/api/users/me'),

  createRoom: () => request<{ data: Room }>('/api/rooms', { method: 'POST' }),

  getRoom: (code: string) => request<{ data: Room }>(`/api/rooms/${code}`),

  joinRoom: (code: string) => request<{ data: Room }>(`/api/rooms/${code}/join`, { method: 'POST' }),

  updateRoomSettings: (code: string, settings: Record<string, unknown>) =>
    request<{ data: Room }>(`/api/rooms/${code}/settings`, {
      method: 'PATCH',
      body: JSON.stringify(settings),
    }),

  kickPlayer: (code: string, userId: string) =>
    request<void>(`/api/rooms/${code}/players/${userId}`, { method: 'DELETE' }),

  getLeaderboard: () => request<{ data: LeaderboardEntry[] }>('/api/games/leaderboard'),

  getGameHistory: () => request<{ data: unknown[] }>('/api/games/history'),

  joinMatchmaking: () => request<void>('/api/matchmaking/queue', { method: 'POST' }),

  leaveMatchmaking: () => request<void>('/api/matchmaking/queue', { method: 'DELETE' }),

  getQueueStatus: () => request<{ in_queue: boolean; queue_size: number }>('/api/matchmaking/status'),
};
