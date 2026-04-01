import { motion } from 'framer-motion';
import type { Card as CardType } from '../types';
import { getSuitSymbol, getSuitColor, getRankLabel } from '../lib/cardUtils';

interface CardProps {
  card: CardType;
  selected?: boolean;
  playable?: boolean;
  faceDown?: boolean;
  small?: boolean;
  onClick?: () => void;
}

export default function Card({ card, selected, playable, faceDown, small, onClick }: CardProps) {
  const w = small ? 48 : 72;
  const h = small ? 68 : 100;
  const fontSize = small ? 11 : 16;
  const suitSize = small ? 14 : 20;

  if (faceDown) {
    return (
      <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className="drop-shadow-md">
        <rect x="1" y="1" width={w - 2} height={h - 2} rx="6" fill="#1a5c2e" stroke="#0d3d1a" strokeWidth="1.5" />
        <rect x="4" y="4" width={w - 8} height={h - 8} rx="4" fill="none" stroke="#2d8a47" strokeWidth="0.5" />
        <text x={w / 2} y={h / 2} textAnchor="middle" dominantBaseline="middle" fill="#2d8a47" fontSize={suitSize} fontWeight="bold">♠</text>
      </svg>
    );
  }

  const color = getSuitColor(card.suit);
  const symbol = getSuitSymbol(card.suit);
  const label = getRankLabel(card.rank);

  return (
    <motion.div
      whileHover={playable ? { y: -8, scale: 1.05 } : undefined}
      whileTap={playable ? { scale: 0.95 } : undefined}
      animate={selected ? { y: -12, boxShadow: '0 0 12px #d4a843' } : { y: 0 }}
      onClick={onClick}
      className={`inline-block cursor-${playable ? 'pointer' : 'default'} select-none`}
      style={{ width: w, height: h }}
    >
      <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className="drop-shadow-lg">
        {/* Card background */}
        <rect x="1" y="1" width={w - 2} height={h - 2} rx="6"
          fill="white"
          stroke={selected ? '#d4a843' : '#d1d5db'}
          strokeWidth={selected ? 2 : 1}
        />
        {/* Top-left rank + suit */}
        <text x="6" y={fontSize + 4} fill={color} fontSize={fontSize} fontWeight="bold" fontFamily="Inter, sans-serif">
          {label}
        </text>
        <text x="6" y={fontSize + 4 + suitSize} fill={color} fontSize={suitSize}>
          {symbol}
        </text>
        {/* Center suit */}
        <text x={w / 2} y={h / 2 + 4} textAnchor="middle" dominantBaseline="middle" fill={color} fontSize={suitSize * 1.8}>
          {symbol}
        </text>
        {/* Bottom-right rank + suit (rotated) */}
        <g transform={`rotate(180, ${w / 2}, ${h / 2})`}>
          <text x="6" y={fontSize + 4} fill={color} fontSize={fontSize} fontWeight="bold" fontFamily="Inter, sans-serif">
            {label}
          </text>
          <text x="6" y={fontSize + 4 + suitSize} fill={color} fontSize={suitSize}>
            {symbol}
          </text>
        </g>
        {/* Playable glow */}
        {playable && !selected && (
          <rect x="1" y="1" width={w - 2} height={h - 2} rx="6" fill="none" stroke="#2d8a47" strokeWidth="1.5" opacity="0.6" />
        )}
      </svg>
    </motion.div>
  );
}
