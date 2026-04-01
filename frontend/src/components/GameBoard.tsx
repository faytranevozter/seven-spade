import { useGameStore } from '../stores/gameStore';
import PlayerHand from './PlayerHand';
import PlayerSlot from './PlayerSlot';
import SuitSequenceDisplay from './SuitSequence';
import GameLog from './GameLog';
import GameOver from './GameOver';
import { useNavigate } from 'react-router-dom';
import { useMemo } from 'react';

const POSITIONS = ['bottom', 'left', 'top', 'right'] as const;

export default function GameBoard() {
  const { gameState, startGame, error } = useGameStore();
  const navigate = useNavigate();

  const playerNames = useMemo(() => {
    if (!gameState) return {};
    const names: Record<number, string> = {};
    gameState.players.forEach((p) => {
      names[p.seat] = p.display_name;
    });
    return names;
  }, [gameState]);

  // Reorder players so that "me" is always at bottom
  const orderedPlayers = useMemo(() => {
    if (!gameState) return [];
    const mySeat = gameState.my_seat;
    const players = [...gameState.players];
    // Sort so seat order is: mySeat, mySeat+1, mySeat+2, mySeat+3 (mod 4)
    players.sort((a, b) => {
      const orderA = (a.seat - mySeat + 4) % 4;
      const orderB = (b.seat - mySeat + 4) % 4;
      return orderA - orderB;
    });
    return players;
  }, [gameState]);

  if (!gameState) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4">
        <div className="animate-pulse text-gray-400">Waiting for game state...</div>
      </div>
    );
  }

  const isMyTurn = gameState.current_turn === gameState.my_seat;

  return (
    <div className="relative w-full h-full min-h-[600px]">
      {/* Error banner */}
      {error && (
        <div className="absolute top-2 left-1/2 -translate-x-1/2 z-50 bg-red-600/90 text-white px-4 py-2 rounded-lg text-sm">
          {error}
        </div>
      )}

      {/* Player slots around the table */}
      <div className="absolute inset-0 m-8">
        {orderedPlayers.map((player, i) => (
          <PlayerSlot
            key={player.seat}
            player={player}
            position={POSITIONS[i]}
            isMe={player.seat === gameState.my_seat}
          />
        ))}
      </div>

      {/* Center: sequences */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-72">
        <SuitSequenceDisplay sequences={gameState.sequences} aceDirection={gameState.ace_direction} />
      </div>

      {/* Turn indicator */}
      <div className="absolute top-4 left-1/2 -translate-x-1/2 text-center">
        <div className={`text-lg font-semibold ${isMyTurn ? 'text-[#d4a843]' : 'text-gray-400'}`}>
          {isMyTurn
            ? gameState.must_face_down
              ? '⚠️ No valid moves — select cards to face down'
              : '🎯 Your turn — play a card!'
            : `Waiting for ${playerNames[gameState.current_turn] || 'opponent'}...`}
        </div>
      </div>

      {/* My hand at bottom */}
      <div className="absolute bottom-4 left-1/2 -translate-x-1/2 max-w-2xl w-full px-4">
        <PlayerHand
          cards={gameState.hand}
          validMoves={gameState.valid_moves}
          mustFaceDown={gameState.must_face_down}
          isMyTurn={isMyTurn}
        />
      </div>

      {/* Game log on the right */}
      <div className="absolute top-4 right-4 w-64">
        <GameLog moves={gameState.move_history} playerNames={playerNames} />
      </div>

      {/* Game over overlay */}
      {gameState.status === 'finished' && gameState.results && (
        <GameOver
          results={gameState.results}
          playerNames={playerNames}
          onPlayAgain={startGame}
          onLeave={() => navigate('/dashboard')}
        />
      )}
    </div>
  );
}
