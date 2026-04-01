import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useGameStore } from '../stores/gameStore';
import GameBoard from '../components/GameBoard';

export default function GamePage() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const { roomCode, connectToRoom, disconnect } = useGameStore();

  useEffect(() => {
    if (!code) {
      navigate('/dashboard');
      return;
    }

    if (roomCode !== code) {
      connectToRoom(code);
    }

    return () => {
      // Don't disconnect on unmount — user might be navigating temporarily
    };
  }, [code, roomCode, connectToRoom, navigate]);

  return (
    <div className="h-screen w-screen overflow-hidden relative bg-gradient-to-br from-[#0d1b0f] to-[#0a1a0c]">
      {/* Header */}
      <div className="absolute top-2 left-4 z-20 flex items-center gap-4">
        <button
          onClick={() => {
            disconnect();
            navigate('/dashboard');
          }}
          className="text-xs text-gray-400 hover:text-white bg-[#1a2e1d]/80 px-3 py-1.5 rounded-lg"
        >
          ← Leave
        </button>
        <span className="text-xs text-gray-500 font-mono">Room: {code}</span>
      </div>

      {/* Game Board */}
      <div className="h-full pt-10">
        <GameBoard />
      </div>
    </div>
  );
}
