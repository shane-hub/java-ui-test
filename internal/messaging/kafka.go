package messaging

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, key string, value []byte) error
	Close() error
}

type InMemoryPublisher struct {
	events map[string][][]byte
}

func NewKafkaPublisher(_ []string, _ string) (*InMemoryPublisher, error) {
	return &InMemoryPublisher{events: map[string][][]byte{}}, nil
}

func (p *InMemoryPublisher) Publish(_ context.Context, key string, value []byte) error {
	p.events[key] = append(p.events[key], append([]byte(nil), value...))
	return nil
}

func (p *InMemoryPublisher) Close() error { return nil }
