import type { GameState } from '../types';

export type WSMessageType =
  | 'play_card'
  | 'face_down'
  | 'request_state'
  | 'start_game'
  | 'room_chat'
  | 'game_state'
  | 'move_made'
  | 'game_over'
  | 'game_starting'
  | 'turn_change'
  | 'player_joined'
  | 'player_left'
  | 'room_chat_bcast'
  | 'error';

export interface WSMessage {
  type: WSMessageType;
  payload: unknown;
}

type MessageHandler = (msg: WSMessage) => void;

export class GameWebSocket {
  private ws: WebSocket | null = null;
  private handlers: Map<string, MessageHandler[]> = new Map();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private roomCode: string = '';

  connect(roomCode: string) {
    this.roomCode = roomCode;
    const token = localStorage.getItem('token');
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}/ws/room/${roomCode}?token=${token}`;

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.emit({ type: 'request_state', payload: {} });
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        const handlers = this.handlers.get(msg.type) || [];
        handlers.forEach((h) => h(msg));

        // Also fire wildcard handlers
        const wildcardHandlers = this.handlers.get('*') || [];
        wildcardHandlers.forEach((h) => h(msg));
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.roomCode) {
        this.connect(this.roomCode);
      }
    }, 2000);
  }

  on(type: string, handler: MessageHandler) {
    const existing = this.handlers.get(type) || [];
    existing.push(handler);
    this.handlers.set(type, existing);
  }

  off(type: string, handler: MessageHandler) {
    const existing = this.handlers.get(type) || [];
    this.handlers.set(
      type,
      existing.filter((h) => h !== handler)
    );
  }

  emit(msg: WSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  playCard(suit: string, rank: number) {
    this.emit({ type: 'play_card', payload: { suit, rank } });
  }

  faceDown(cards: { suit: string; rank: number }[]) {
    this.emit({ type: 'face_down', payload: { cards } });
  }

  startGame() {
    this.emit({ type: 'start_game', payload: {} });
  }

  sendChat(message: string) {
    this.emit({ type: 'room_chat', payload: { message } });
  }

  requestState() {
    this.emit({ type: 'request_state', payload: {} });
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.roomCode = '';
    this.ws?.close();
    this.ws = null;
    this.handlers.clear();
  }

  get connected() {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Parse game_state payload
export function parseGameState(payload: unknown): GameState | null {
  if (!payload || typeof payload !== 'object') return null;
  return payload as GameState;
}

export const gameWS = new GameWebSocket();
