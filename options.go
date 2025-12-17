package JustGoHTML

import (
	"github.com/MeKo-Christian/JustGoHTML/treebuilder"
)

// config holds parser configuration.
type config struct {
	encoding        string
	fragmentContext *treebuilder.FragmentContext
	iframeSrcdoc    bool
	strict          bool
	collectErrors   bool
}

// newConfig creates a new config with defaults and applies options.
func newConfig(opts ...Option) *config {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Option configures the parser behavior.
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

// WithFragment sets the parsing context for fragment parsing.
// This is typically used internally by ParseFragment.
func WithFragment(tagName string) Option {
	return func(c *config) {
		c.fragmentContext = &treebuilder.FragmentContext{
			TagName:   tagName,
			Namespace: "html",
		}
	}
}

// WithFragmentNS sets the parsing context with a specific namespace.
// Use this for parsing SVG or MathML fragments.
func WithFragmentNS(tagName, namespace string) Option {
	return func(c *config) {
		c.fragmentContext = &treebuilder.FragmentContext{
			TagName:   tagName,
			Namespace: namespace,
		}
	}
}

// WithIframeSrcdoc enables iframe srcdoc parsing mode.
// In this mode, the parser treats the input as the srcdoc attribute value.
func WithIframeSrcdoc() Option {
	return func(c *config) {
		c.iframeSrcdoc = true
	}
}

// WithStrictMode enables strict parsing mode.
// In this mode, the first parse error causes Parse to return an error.
// By default, parse errors are handled according to the HTML5 spec
// and parsing continues.
func WithStrictMode() Option {
	return func(c *config) {
		c.strict = true
	}
}

// WithCollectErrors enables error collection mode.
// Parse errors are collected and returned as a ParseErrors error
// (which can be unwrapped to get individual errors).
// Without this option, parse errors are silently recovered from.
func WithCollectErrors() Option {
	return func(c *config) {
		c.collectErrors = true
	}
}
