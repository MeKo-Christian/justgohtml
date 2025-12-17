// Package stream provides a streaming API for HTML parsing.
package stream

import (
	"github.com/MeKo-Christian/JustGoHTML/encoding"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

// EventType represents the type of streaming event.
type EventType int

// Event types for the streaming API.
const (
	StartTagEvent EventType = iota
	EndTagEvent
	TextEvent
	CommentEvent
	DoctypeEvent
)

// String returns the name of the event type.
func (e EventType) String() string {
	names := [...]string{"StartTag", "EndTag", "Text", "Comment", "Doctype"}
	if int(e) < len(names) {
		return names[e]
	}
	return "Unknown"
}

// Event represents a parsing event in the stream.
type Event struct {
	// Type is the event type.
	Type EventType

	// Name is the tag name (for start/end tags) or DOCTYPE name.
	Name string

	// Attrs contains attributes (for start tags only).
	Attrs map[string]string

	// Data is the text content (for text/comment events).
	Data string

	// For DOCTYPE events
	PublicID string
	SystemID string
}

// Stream returns a channel of parsing events.
// The channel is closed when parsing is complete.
func Stream(html string) <-chan Event {
	ch := make(chan Event)
	go func() {
		defer close(ch)
		streamTokens(html, ch)
	}()
	return ch
}

// StreamBytes returns a channel of parsing events from byte input.
func StreamBytes(html []byte) <-chan Event {
	decoded, _, err := encoding.Decode(html, "")
	if err != nil {
		ch := make(chan Event)
		close(ch)
		return ch
	}
	return Stream(decoded)
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func streamTokens(html string, ch chan<- Event) {
	tok := tokenizer.New(html)

	for {
		token := tok.Next()

		switch token.Type {
		case tokenizer.StartTag:
			ch <- Event{
				Type:  StartTagEvent,
				Name:  token.Name,
				Attrs: token.Attrs,
			}

		case tokenizer.EndTag:
			ch <- Event{
				Type: EndTagEvent,
				Name: token.Name,
			}

		case tokenizer.Character:
			ch <- Event{
				Type: TextEvent,
				Data: token.Data,
			}

		case tokenizer.Comment:
			ch <- Event{
				Type: CommentEvent,
				Data: token.Data,
			}

		case tokenizer.DOCTYPE:
			ch <- Event{
				Type:     DoctypeEvent,
				Name:     token.Name,
				PublicID: ptrToString(token.PublicID),
				SystemID: ptrToString(token.SystemID),
			}

		case tokenizer.EOF:
			return

		case tokenizer.Error:
			// Continue on errors (per HTML5 spec)
			continue
		}
	}
}
