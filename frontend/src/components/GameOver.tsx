import { motion } from 'framer-motion';
import type { PlayerResult } from '../types';

interface GameOverProps {
  results: PlayerResult[];
  playerNames: Record<number, string>;
  onPlayAgain: () => void;
  onLeave: () => void;
}

export default function GameOver({ results, playerNames, onPlayAgain, onLeave }: GameOverProps) {
  const sorted = [...results].sort((a, b) => a.rank - b.rank);

  const medals = ['🥇', '🥈', '🥉', '4th'];

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.8 }}
      animate={{ opacity: 1, scale: 1 }}
      className="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
    >
      <div className="bg-[#1a2e1d] rounded-2xl p-8 max-w-md w-full mx-4 border border-[#2d8a47]/50">
        <h2 className="text-2xl font-bold text-center text-[#d4a843] mb-6">Game Over!</h2>

        <div className="space-y-3 mb-8">
          {sorted.map((r, i) => (
            <motion.div
              key={r.seat}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: i * 0.15 }}
              className={`flex items-center gap-4 p-3 rounded-lg ${
                i === 0 ? 'bg-[#d4a843]/20 ring-1 ring-[#d4a843]' : 'bg-[#0d1b0f]/50'
              }`}
            >
              <span className="text-2xl w-10 text-center">{medals[i]}</span>
              <div className="flex-1">
                <div className="font-semibold">{playerNames[r.seat] || `Player ${r.seat + 1}`}</div>
                <div className="text-sm text-gray-400">
                  {r.penalty_points} penalty point{r.penalty_points !== 1 ? 's' : ''}
                </div>
              </div>
            </motion.div>
          ))}
        </div>

        <div className="flex gap-3">
          <button
            onClick={onPlayAgain}
            className="flex-1 py-3 bg-[#2d8a47] hover:bg-[#1a5c2e] text-white rounded-lg font-semibold transition-colors"
          >
            Play Again
          </button>
          <button
            onClick={onLeave}
            className="flex-1 py-3 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-semibold transition-colors"
          >
            Leave
          </button>
        </div>
      </div>
    </motion.div>
  );
}
