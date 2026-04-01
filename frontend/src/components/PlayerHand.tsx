import { motion, AnimatePresence } from 'framer-motion';
import type { Card as CardType } from '../types';
import Card from './Card';
import { useGameStore } from '../stores/gameStore';

interface PlayerHandProps {
  cards: CardType[];
  validMoves: CardType[];
  mustFaceDown: boolean;
  isMyTurn: boolean;
}

function isCardPlayable(card: CardType, validMoves: CardType[]): boolean {
  return validMoves.some((v) => v.suit === card.suit && v.rank === card.rank);
}

export default function PlayerHand({ cards, validMoves, mustFaceDown, isMyTurn }: PlayerHandProps) {
  const { selectedCards, selectCard, deselectCard, playCard, faceDown } = useGameStore();

  const isSelected = (card: CardType) =>
    selectedCards.some((c) => c.suit === card.suit && c.rank === card.rank);

  const handleCardClick = (card: CardType) => {
    if (!isMyTurn) return;

    if (mustFaceDown) {
      if (isSelected(card)) {
        deselectCard(card);
      } else {
        selectCard(card);
      }
      return;
    }

    if (isCardPlayable(card, validMoves)) {
      playCard(card);
    }
  };

  // Sort hand by suit then rank
  const sorted = [...cards].sort((a, b) => {
    const suitOrder = ['spades', 'hearts', 'diamonds', 'clubs'];
    const si = suitOrder.indexOf(a.suit) - suitOrder.indexOf(b.suit);
    if (si !== 0) return si;
    return a.rank - b.rank;
  });

  return (
    <div className="flex flex-col items-center gap-3">
      <div className="flex flex-wrap justify-center gap-1">
        <AnimatePresence>
          {sorted.map((card) => (
            <motion.div
              key={`${card.suit}-${card.rank}`}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ duration: 0.2 }}
            >
              <Card
                card={card}
                selected={isSelected(card)}
                playable={isMyTurn && (mustFaceDown || isCardPlayable(card, validMoves))}
                onClick={() => handleCardClick(card)}
              />
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {isMyTurn && mustFaceDown && selectedCards.length > 0 && (
        <motion.button
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          className="px-6 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-semibold transition-colors"
          onClick={faceDown}
        >
          Face Down ({selectedCards.length} card{selectedCards.length > 1 ? 's' : ''})
        </motion.button>
      )}
    </div>
  );
}
