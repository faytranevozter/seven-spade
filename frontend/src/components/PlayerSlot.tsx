import { motion } from 'framer-motion';
import type { PlayerView } from '../types';

interface PlayerSlotProps {
  player: PlayerView;
  position: 'bottom' | 'left' | 'top' | 'right';
  isMe: boolean;
}

const positionStyles: Record<string, string> = {
  bottom: 'bottom-0 left-1/2 -translate-x-1/2',
  top: 'top-0 left-1/2 -translate-x-1/2',
  left: 'left-0 top-1/2 -translate-y-1/2',
  right: 'right-0 top-1/2 -translate-y-1/2',
};

export default function PlayerSlot({ player, position, isMe }: PlayerSlotProps) {
  return (
    <div className={`absolute ${positionStyles[position]} z-10`}>
      <motion.div
        animate={player.is_current_turn ? { scale: [1, 1.05, 1] } : {}}
        transition={{ duration: 1, repeat: Infinity }}
        className={`flex flex-col items-center gap-1 px-4 py-2 rounded-xl
          ${player.is_current_turn ? 'bg-[#2d8a47]/30 ring-2 ring-[#d4a843]' : 'bg-[#1a2e1d]/60'}
          ${isMe ? 'ring-1 ring-[#2d8a47]' : ''}
        `}
      >
        <div className="flex items-center gap-2">
          {player.is_bot && <span className="text-xs">🤖</span>}
          <span className={`text-sm font-semibold ${isMe ? 'text-[#d4a843]' : 'text-gray-200'}`}>
            {player.display_name}
          </span>
        </div>
        <div className="flex items-center gap-3 text-xs text-gray-400">
          <span>🃏 {player.card_count}</span>
          {player.face_down_count > 0 && (
            <span className="text-red-400">⬇ {player.face_down_count}</span>
          )}
        </div>
      </motion.div>
    </div>
  );
}
