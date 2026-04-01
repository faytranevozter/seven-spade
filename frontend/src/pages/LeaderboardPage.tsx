import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import type { LeaderboardEntry } from '../types';

export default function LeaderboardPage() {
  const navigate = useNavigate();
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getLeaderboard()
      .then((res) => setEntries(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-2xl mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-gray-400 hover:text-white mb-6 text-sm">
          ← Back to Dashboard
        </button>

        <h1 className="text-2xl font-bold text-[#d4a843] mb-6">🏆 Leaderboard</h1>

        {loading ? (
          <div className="text-gray-400 animate-pulse">Loading...</div>
        ) : entries.length === 0 ? (
          <div className="text-gray-500">No games played yet. Be the first!</div>
        ) : (
          <div className="bg-[#1a2e1d] rounded-xl border border-[#2d8a47]/30 overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[#2d8a47]/20 text-gray-400">
                  <th className="px-4 py-3 text-left">#</th>
                  <th className="px-4 py-3 text-left">Player</th>
                  <th className="px-4 py-3 text-right">ELO</th>
                  <th className="px-4 py-3 text-right">Games</th>
                  <th className="px-4 py-3 text-right">Wins</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((entry, i) => (
                  <tr key={entry.user_id} className="border-b border-[#2d8a47]/10 hover:bg-[#0d1b0f]/30">
                    <td className="px-4 py-3">
                      {i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : i + 1}
                    </td>
                    <td className="px-4 py-3 font-semibold">{entry.display_name}</td>
                    <td className="px-4 py-3 text-right text-[#d4a843]">{entry.elo_rating}</td>
                    <td className="px-4 py-3 text-right text-gray-400">{entry.total_games}</td>
                    <td className="px-4 py-3 text-right text-green-400">{entry.wins}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
