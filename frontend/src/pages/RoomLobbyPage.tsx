import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { api } from '../lib/api';
import { useAuthStore } from '../stores/authStore';
import { useGameStore } from '../stores/gameStore';
import type { Room } from '../types';

export default function RoomLobbyPage() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { connectToRoom } = useGameStore();
  const [room, setRoom] = useState<Room | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!code) return;
    const load = async () => {
      try {
        const res = await api.getRoom(code);
        setRoom(res.data);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Room not found');
      }
    };
    load();
    const interval = setInterval(load, 3000);
    return () => clearInterval(interval);
  }, [code]);

  const isHost = room?.host_user_id === user?.id;

  const handleStart = () => {
    if (!code) return;
    connectToRoom(code);
    navigate(`/game/${code}`);
  };

  const handleJoinAndPlay = () => {
    if (!code) return;
    connectToRoom(code);
    navigate(`/game/${code}`);
  };

  const handleUpdateSettings = async (settings: Record<string, unknown>) => {
    if (!code) return;
    try {
      const res = await api.updateRoomSettings(code, settings);
      setRoom(res.data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update settings');
    }
  };

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={() => navigate('/dashboard')} className="text-[#d4a843] hover:underline">
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  if (!room) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-gray-400">Loading room...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-2xl mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-gray-400 hover:text-white mb-6 text-sm">
          ← Back to Dashboard
        </button>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-[#1a2e1d] rounded-2xl p-8 border border-[#2d8a47]/30"
        >
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold">Room Lobby</h2>
            <div className="bg-[#0d1b0f] px-4 py-2 rounded-lg font-mono text-2xl tracking-widest text-[#d4a843]">
              {room.code}
            </div>
          </div>

          {/* Players */}
          <div className="mb-6">
            <h3 className="text-sm text-gray-400 mb-3">Players ({room.players?.length || 0}/4)</h3>
            <div className="grid grid-cols-2 gap-3">
              {[0, 1, 2, 3].map((seat) => {
                const player = room.players?.find((p) => p.seat === seat);
                return (
                  <div
                    key={seat}
                    className={`p-3 rounded-lg border ${
                      player
                        ? 'bg-[#0d1b0f] border-[#2d8a47]/40'
                        : 'bg-[#0d1b0f]/30 border-dashed border-gray-600'
                    }`}
                  >
                    {player ? (
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="font-semibold text-sm">
                            {player.is_bot ? `🤖 Bot (${player.bot_difficulty})` : player.user?.display_name || 'Player'}
                          </div>
                          <div className="text-xs text-gray-400">Seat {seat + 1}</div>
                        </div>
                        {isHost && player.user_id !== user?.id && !player.is_bot && (
                          <button
                            onClick={() => api.kickPlayer(room.code, player.user_id)}
                            className="text-xs text-red-400 hover:text-red-300"
                          >
                            Kick
                          </button>
                        )}
                      </div>
                    ) : (
                      <div className="text-gray-500 text-sm text-center">Empty Seat</div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Settings (host only) */}
          {isHost && (
            <div className="mb-6 p-4 bg-[#0d1b0f]/50 rounded-lg">
              <h3 className="text-sm text-gray-400 mb-3">Room Settings</h3>
              <div className="flex flex-wrap gap-4">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={room.bot_enabled}
                    onChange={(e) => handleUpdateSettings({ bot_enabled: e.target.checked })}
                    className="rounded"
                  />
                  Enable Bots
                </label>
                {room.bot_enabled && (
                  <div className="flex items-center gap-2 text-sm">
                    <span>Bot Count:</span>
                    <select
                      value={room.bot_count}
                      onChange={(e) => handleUpdateSettings({ bot_count: parseInt(e.target.value) })}
                      className="bg-[#0d1b0f] border border-[#2d8a47]/40 rounded px-2 py-1 text-sm"
                    >
                      {[1, 2, 3].map((n) => (
                        <option key={n} value={n}>{n}</option>
                      ))}
                    </select>
                  </div>
                )}
                <div className="flex items-center gap-2 text-sm">
                  <span>Turn Timer:</span>
                  <select
                    value={room.turn_timer}
                    onChange={(e) => handleUpdateSettings({ turn_timer: parseInt(e.target.value) })}
                    className="bg-[#0d1b0f] border border-[#2d8a47]/40 rounded px-2 py-1 text-sm"
                  >
                    <option value={0}>No Limit</option>
                    <option value={15}>15s</option>
                    <option value={30}>30s</option>
                    <option value={60}>60s</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3">
            {isHost ? (
              <button
                onClick={handleStart}
                className="flex-1 py-3 bg-[#d4a843] hover:bg-[#b88f35] text-black rounded-lg font-semibold transition-colors"
              >
                Start Game
              </button>
            ) : (
              <button
                onClick={handleJoinAndPlay}
                className="flex-1 py-3 bg-[#2d8a47] hover:bg-[#1a5c2e] text-white rounded-lg font-semibold transition-colors"
              >
                Ready / Connect
              </button>
            )}
          </div>

          {/* Share link */}
          <div className="mt-4 text-center">
            <p className="text-xs text-gray-500">
              Share this code with friends: <span className="text-[#d4a843] font-mono">{room.code}</span>
            </p>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
