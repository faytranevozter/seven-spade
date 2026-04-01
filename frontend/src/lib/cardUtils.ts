import type { Suit, Rank } from '../types';

const SUIT_SYMBOLS: Record<Suit, string> = {
  spades: '♠',
  hearts: '♥',
  diamonds: '♦',
  clubs: '♣',
};

const SUIT_COLORS: Record<Suit, string> = {
  spades: '#1f2937',
  clubs: '#1f2937',
  hearts: '#dc2626',
  diamonds: '#dc2626',
};

const RANK_LABELS: Record<Rank, string> = {
  1: 'A', 2: '2', 3: '3', 4: '4', 5: '5', 6: '6', 7: '7',
  8: '8', 9: '9', 10: '10', 11: 'J', 12: 'Q', 13: 'K',
};

export function getSuitSymbol(suit: Suit): string {
  return SUIT_SYMBOLS[suit];
}

export function getSuitColor(suit: Suit): string {
  return SUIT_COLORS[suit];
}

export function getRankLabel(rank: Rank): string {
  return RANK_LABELS[rank];
}
