// Package stream provides options for configuring streaming HTML parsing.
package stream

// config holds stream configuration.
type config struct {
	encoding string
}

// newConfig creates a new config with defaults and applies options.
func newConfig(opts ...Option) *config {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Option configures the streaming parser behavior.
type Option func(*config)

// WithEncoding sets the character encoding to use for parsing.
// This overrides automatic encoding detection.
//
// Common values: "utf-8", "windows-1252", "iso-8859-1"
func WithEncoding(enc string) Option {
	return func(c *config) {
		c.encoding = enc
	}
}
