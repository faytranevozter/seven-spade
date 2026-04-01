import { create } from 'zustand';
import type { GameState, Card } from '../types';
import { gameWS, parseGameState } from '../lib/ws';

interface GameStore {
  gameState: GameState | null;
  selectedCards: Card[];
  roomCode: string | null;
  chatMessages: { user: string; message: string }[];
  error: string | null;

  connectToRoom: (code: string) => void;
  disconnect: () => void;
  selectCard: (card: Card) => void;
  deselectCard: (card: Card) => void;
  clearSelection: () => void;
  playCard: (card: Card) => void;
  faceDown: () => void;
  startGame: () => void;
  sendChat: (message: string) => void;
}

export const useGameStore = create<GameStore>((set, get) => ({
  gameState: null,
  selectedCards: [],
  roomCode: null,
  chatMessages: [],
  error: null,

  connectToRoom: (code: string) => {
    set({ roomCode: code, error: null });
    gameWS.disconnect();

    gameWS.on('game_state', (msg) => {
      const state = parseGameState(msg.payload);
      if (state) set({ gameState: state, error: null });
    });

    gameWS.on('move_made', () => {
      gameWS.requestState();
    });

    gameWS.on('turn_change', () => {
      gameWS.requestState();
    });

    gameWS.on('game_over', (msg) => {
      const state = parseGameState(msg.payload);
      if (state) set({ gameState: state });
    });

    gameWS.on('game_starting', () => {
      set({ gameState: null, selectedCards: [] });
    });

    gameWS.on('room_chat_bcast', (msg) => {
      const payload = msg.payload as { display_name: string; message: string };
      set((s) => ({
        chatMessages: [...s.chatMessages, { user: payload.display_name, message: payload.message }],
      }));
    });

    gameWS.on('error', (msg) => {
      const payload = msg.payload as { message: string };
      set({ error: payload.message });
    });

    gameWS.connect(code);
  },

  disconnect: () => {
    gameWS.disconnect();
    set({ gameState: null, roomCode: null, selectedCards: [], chatMessages: [], error: null });
  },

  selectCard: (card: Card) => {
    set((s) => ({ selectedCards: [...s.selectedCards, card] }));
  },

  deselectCard: (card: Card) => {
    set((s) => ({
      selectedCards: s.selectedCards.filter(
        (c) => !(c.suit === card.suit && c.rank === card.rank)
      ),
    }));
  },

  clearSelection: () => set({ selectedCards: [] }),

  playCard: (card: Card) => {
    gameWS.playCard(card.suit, card.rank);
    set({ selectedCards: [] });
  },

  faceDown: () => {
    const { selectedCards } = get();
    if (selectedCards.length > 0) {
      gameWS.faceDown(selectedCards.map((c) => ({ suit: c.suit, rank: c.rank })));
      set({ selectedCards: [] });
    }
  },

  startGame: () => {
    gameWS.startGame();
  },

  sendChat: (message: string) => {
    gameWS.sendChat(message);
  },
}));
