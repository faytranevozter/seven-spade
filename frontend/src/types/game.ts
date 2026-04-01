export type Suit = 'spades' | 'hearts' | 'diamonds' | 'clubs';
export type Rank = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13;

export interface Card {
  suit: Suit;
  rank: Rank;
}

export type AceDirection = 'undecided' | 'low' | 'high';
export type GameStatus = 'waiting' | 'playing' | 'finished';
export type MoveType = 'play' | 'face_down';

export interface SuitSequence {
  suit: Suit;
  low: number;
  high: number;
  closed: boolean;
}

export interface PlayerResult {
  seat: number;
  penalty_points: number;
  rank: number;
  face_down_cards: Card[];
}

export interface Move {
  seat: number;
  move_num: number;
  type: MoveType;
  card: Card;
}

export interface PlayerView {
  seat: number;
  user_id: string;
  display_name: string;
  card_count: number;
  face_down_count: number;
  is_bot: boolean;
  is_current_turn: boolean;
}

export interface GameState {
  hand: Card[];
  sequences: SuitSequence[];
  players: PlayerView[];
  current_turn: number;
  ace_direction: AceDirection;
  status: GameStatus;
  valid_moves: Card[];
  must_face_down: boolean;
  move_history: Move[];
  results?: PlayerResult[];
  my_seat: number;
}

export interface User {
  id: string;
  display_name: string;
  email: string;
  avatar_url: string;
  elo_rating: number;
  stats?: PlayerStats;
}

export interface PlayerStats {
  elo_rating: number;
  total_games: number;
  wins: number;
  losses: number;
  total_penalty_points: number;
}

export interface Room {
  id: string;
  code: string;
  host_user_id: string;
  status: string;
  bot_enabled: boolean;
  bot_count: number;
  turn_timer: number;
  players: RoomPlayer[];
}

export interface RoomPlayer {
  user_id: string;
  seat: number;
  is_bot: boolean;
  bot_difficulty: string;
  user?: User;
}

export interface LeaderboardEntry {
  user_id: string;
  display_name: string;
  elo_rating: number;
  total_games: number;
  wins: number;
}
