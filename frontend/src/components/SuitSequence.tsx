import type { SuitSequence as SuitSeqType, AceDirection } from '../types';
import { getSuitSymbol, getSuitColor, getRankLabel } from '../lib/cardUtils';
import type { Rank } from '../types';

interface SuitSequenceProps {
  sequences: SuitSeqType[];
  aceDirection: AceDirection;
}

export default function SuitSequenceDisplay({ sequences, aceDirection }: SuitSequenceProps) {
  const suitOrder = ['spades', 'hearts', 'diamonds', 'clubs'] as const;

  return (
    <div className="grid grid-cols-2 gap-3 p-4 bg-[#1a2e1d] rounded-xl border border-[#2d8a47]/30">
      <div className="col-span-2 text-center text-sm text-gray-400 mb-1">
        Table Sequences
        {aceDirection !== 'undecided' && (
          <span className="ml-2 text-[#d4a843]">
            (Ace goes {aceDirection === 'low' ? '↓ Low' : '↑ High'})
          </span>
        )}
      </div>
      {suitOrder.map((suit) => {
        const seq = sequences.find((s) => s.suit === suit);
        const symbol = getSuitSymbol(suit);
        const color = getSuitColor(suit);

        return (
          <div
            key={suit}
            className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[#0d1b0f]/50"
          >
            <span style={{ color, fontSize: 24 }}>{symbol}</span>
            <div className="flex-1 text-sm">
              {seq ? (
                <span className="font-mono">
                  {getRankLabel(seq.low as Rank)} — {getRankLabel(seq.high as Rank)}
                  {seq.closed && <span className="ml-1 text-red-400">✕</span>}
                </span>
              ) : (
                <span className="text-gray-500">—</span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
