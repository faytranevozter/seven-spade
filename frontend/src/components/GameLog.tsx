import type { Move } from '../types';
import { getSuitSymbol, getRankLabel } from '../lib/cardUtils';
import type { Rank, Suit } from '../types';
import { useEffect, useRef } from 'react';

interface GameLogProps {
  moves: Move[];
  playerNames: Record<number, string>;
}

export default function GameLog({ moves, playerNames }: GameLogProps) {
  const endRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [moves.length]);

  const last20 = moves.slice(-20);

  return (
    <div className="h-48 overflow-y-auto bg-[#0d1b0f]/80 rounded-lg p-3 text-xs font-mono space-y-1">
      {last20.map((m, i) => (
        <div key={i} className="text-gray-400">
          <span className="text-gray-500">{m.move_num}.</span>{' '}
          <span className="text-gray-300">{playerNames[m.seat] || `Seat ${m.seat}`}</span>{' '}
          {m.type === 'play' ? (
            <span>
              plays{' '}
              <span style={{ color: m.card.suit === 'hearts' || m.card.suit === 'diamonds' ? '#dc2626' : '#e2e8f0' }}>
                {getRankLabel(m.card.rank as Rank)}{getSuitSymbol(m.card.suit as Suit)}
              </span>
            </span>
          ) : (
            <span className="text-red-400">faces down a card</span>
          )}
        </div>
      ))}
      <div ref={endRef} />
    </div>
  );
}
