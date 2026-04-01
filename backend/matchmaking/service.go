package matchmaking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// QueueEntry represents a player in the matchmaking queue
type QueueEntry struct {
	UserID      uuid.UUID
	DisplayName string
	EloRating   int
	JoinedAt    time.Time
}

// MatchResult represents a found match
type MatchResult struct {
	Players  []QueueEntry
	RoomCode string
}

// Service manages the matchmaking queue
type Service struct {
	queue   []QueueEntry
	mu      sync.Mutex

	// Callback when a match is found
	OnMatchFound func(match MatchResult)
}

func NewService() *Service {
	return &Service{
		queue: make([]QueueEntry, 0),
	}
}

// JoinQueue adds a player to the matchmaking queue
func (s *Service) JoinQueue(entry QueueEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if player is already in queue
	for _, e := range s.queue {
		if e.UserID == entry.UserID {
			return fmt.Errorf("already in matchmaking queue")
		}
	}

	entry.JoinedAt = time.Now()
	s.queue = append(s.queue, entry)

	logrus.Infof("Player %s joined matchmaking queue (ELO: %d, queue size: %d)",
		entry.DisplayName, entry.EloRating, len(s.queue))

	return nil
}

// LeaveQueue removes a player from the queue
func (s *Service) LeaveQueue(userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.queue {
		if e.UserID == userID {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			logrus.Infof("Player %s left matchmaking queue", e.DisplayName)
			return
		}
	}
}

// QueueSize returns the current queue size
func (s *Service) QueueSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}

// StartMatchmaking begins the background matchmaking loop
func (s *Service) StartMatchmaking(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tryMatch()
		}
	}
}

func (s *Service) tryMatch() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) < 4 {
		return
	}

	// Sort-of ELO-based matching: expand range over time
	// For each player, find 3 others within ELO range
	for i := 0; i < len(s.queue); i++ {
		player := s.queue[i]
		waitTime := time.Since(player.JoinedAt)

		// ELO range expands over time: starts at 200, expands by 100 every 10 seconds
		eloRange := 200 + int(waitTime.Seconds()/10)*100

		matched := []int{i}
		for j := 0; j < len(s.queue) && len(matched) < 4; j++ {
			if i == j {
				continue
			}
			other := s.queue[j]
			if abs(player.EloRating-other.EloRating) <= eloRange {
				matched = append(matched, j)
			}
		}

		if len(matched) == 4 {
			// Extract matched players
			matchPlayers := make([]QueueEntry, 4)
			for k, idx := range matched {
				matchPlayers[k] = s.queue[idx]
			}

			// Remove from queue (in reverse order to preserve indices)
			for k := len(matched) - 1; k >= 0; k-- {
				idx := matched[k]
				s.queue = append(s.queue[:idx], s.queue[idx+1:]...)
			}

			logrus.Infof("Match found! Players: %v", matchPlayerNames(matchPlayers))

			if s.OnMatchFound != nil {
				go s.OnMatchFound(MatchResult{Players: matchPlayers})
			}

			return
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func matchPlayerNames(players []QueueEntry) []string {
	names := make([]string, len(players))
	for i, p := range players {
		names[i] = p.DisplayName
	}
	return names
}

// --- ELO Calculation ---

// UpdateELO calculates new ELO ratings for a 4-player game.
// Results should be sorted by rank (1st place = index 0).
func UpdateELO(ratings []int, ranks []int) []int {
	k := 32.0 // K-factor
	n := len(ratings)
	newRatings := make([]int, n)
	copy(newRatings, ratings)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Expected scores
			ea := 1.0 / (1.0 + pow10(float64(ratings[j]-ratings[i])/400.0))
			eb := 1.0 - ea

			// Actual scores based on ranks
			var sa, sb float64
			if ranks[i] < ranks[j] {
				sa, sb = 1.0, 0.0
			} else if ranks[i] > ranks[j] {
				sa, sb = 0.0, 1.0
			} else {
				sa, sb = 0.5, 0.5
			}

			// Update
			delta := k / float64(n-1)
			newRatings[i] += int(delta * (sa - ea))
			newRatings[j] += int(delta * (sb - eb))
		}
	}

	return newRatings
}

func pow10(x float64) float64 {
	result := 1.0
	base := 10.0
	if x < 0 {
		base = 0.1
		x = -x
	}
	// Simple approximation using repeated multiplication
	whole := int(x)
	for range whole {
		result *= base
	}
	// Fractional part approximation (linear interpolation)
	frac := x - float64(whole)
	if frac > 0 {
		result *= 1 + frac*(base-1)
	}
	return result
}
