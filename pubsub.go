package wattpilot

import "sync"

type Pubsub struct {
	mu     sync.RWMutex
	subs   map[string][]chan interface{}
	closed bool
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[string][]chan interface{})
	return ps
}

func (ps *Pubsub) IsEmpty() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return len(ps.subs) == 0
}

func (ps *Pubsub) Subscribe(topic string) <-chan interface{} {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan interface{}, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *Pubsub) Publish(topic string, msg interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}
	for _, ch := range ps.subs[topic] {
		go func(ch chan interface{}) {
			ch <- msg
		}(ch)
	}
}
func (ps *Pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}
