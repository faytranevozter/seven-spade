import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useAuthStore } from '../stores/authStore';
import { api } from '../lib/api';

export default function DashboardPage() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [joinCode, setJoinCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleCreateRoom = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await api.createRoom();
      navigate(`/room/${res.data.code}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create room');
    } finally {
      setLoading(false);
    }
  };

  const handleJoinRoom = async () => {
    if (!joinCode.trim()) return;
    setLoading(true);
    setError('');
    try {
      await api.joinRoom(joinCode.trim().toUpperCase());
      navigate(`/room/${joinCode.trim().toUpperCase()}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to join room');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen p-6">
      {/* Header */}
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-3">
            <span className="text-3xl">♠</span>
            <h1 className="text-2xl font-bold text-[#d4a843]">Seven Spade</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <div className="text-sm font-semibold">{user?.display_name}</div>
              <div className="text-xs text-gray-400">ELO: {user?.elo_rating || 1000}</div>
            </div>
            <button onClick={logout} className="text-sm text-gray-400 hover:text-white transition-colors">
              Sign Out
            </button>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-600/20 border border-red-600/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div className="grid md:grid-cols-2 gap-6">
          {/* Create Room */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-[#1a2e1d] rounded-xl p-6 border border-[#2d8a47]/30"
          >
            <h2 className="text-xl font-semibold mb-4">Create a Room</h2>
            <p className="text-gray-400 text-sm mb-6">
              Start a new game room and invite friends with the room code.
            </p>
            <button
              onClick={handleCreateRoom}
              disabled={loading}
              className="w-full py-3 bg-[#2d8a47] hover:bg-[#1a5c2e] disabled:opacity-50 text-white rounded-lg font-semibold transition-colors"
            >
              {loading ? 'Creating...' : 'Create Room'}
            </button>
          </motion.div>

          {/* Join Room */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="bg-[#1a2e1d] rounded-xl p-6 border border-[#2d8a47]/30"
          >
            <h2 className="text-xl font-semibold mb-4">Join a Room</h2>
            <p className="text-gray-400 text-sm mb-4">
              Enter the room code shared by your friend.
            </p>
            <div className="flex gap-2">
              <input
                type="text"
                value={joinCode}
                onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                placeholder="ABCDEF"
                maxLength={6}
                className="flex-1 px-4 py-3 bg-[#0d1b0f] border border-[#2d8a47]/40 rounded-lg text-center text-lg font-mono tracking-widest focus:outline-none focus:ring-2 focus:ring-[#d4a843]"
                onKeyDown={(e) => e.key === 'Enter' && handleJoinRoom()}
              />
              <button
                onClick={handleJoinRoom}
                disabled={loading || !joinCode.trim()}
                className="px-6 py-3 bg-[#d4a843] hover:bg-[#b88f35] disabled:opacity-50 text-black rounded-lg font-semibold transition-colors"
              >
                Join
              </button>
            </div>
          </motion.div>
        </div>

        {/* Quick actions */}
        <div className="mt-6 flex gap-4">
          <button
            onClick={() => navigate('/leaderboard')}
            className="px-4 py-2 bg-[#1a2e1d] border border-[#2d8a47]/30 rounded-lg text-sm text-gray-300 hover:text-white transition-colors"
          >
            🏆 Leaderboard
          </button>
        </div>
      </div>
    </div>
  );
}
