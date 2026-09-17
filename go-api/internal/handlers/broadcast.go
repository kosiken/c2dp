package handlers

import (
	"sync"

	"c2dp/go-api/internal/models"
)

// Broadcaster fans out newly created posts to live subscribers, such as the
// presentation's real-time slide.
type Broadcaster struct {
	mu   sync.Mutex
	subs map[chan models.Post]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[chan models.Post]struct{})}
}

func (b *Broadcaster) Subscribe() chan models.Post {
	ch := make(chan models.Post, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan models.Post) {
	b.mu.Lock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Broadcaster) Publish(post models.Post) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- post:
		default:
			// Drop the event for a slow subscriber rather than blocking uploads.
		}
	}
}
