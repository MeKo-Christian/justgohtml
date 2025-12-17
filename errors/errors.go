// Package errors defines parse errors for the HTML5 parser.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotImplemented is returned when a feature is not yet implemented.
var ErrNotImplemented = errors.New("not implemented")

// ParseError represents a single parse error with location information.
type ParseError struct {
	// Code is the error code (e.g., "unexpected-null-character").
	// These codes follow the WHATWG HTML5 specification.
	Code string

	// Message is a human-readable error message.
	Message string

	// Line is the 1-based line number where the error occurred.
	Line int

	// Column is the 1-based column number where the error occurred.
	Column int
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("%s at %d:%d: %s", e.Code, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ParseErrors is a collection of parse errors.
// It implements the error interface so it can be returned from Parse.
type ParseErrors []*ParseError

// Error implements the error interface.
func (e ParseErrors) Error() string {
	if len(e) == 0 {
		return "no parse errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d parse errors:\n", len(e)))
	for i, err := range e {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Unwrap returns the underlying errors for errors.Is/As support.
func (e ParseErrors) Unwrap() []error {
	errs := make([]error, len(e))
	for i, err := range e {
		errs[i] = err
	}
	return errs
}

// SelectorError represents an error in CSS selector parsing.
type SelectorError struct {
	// Selector is the original selector string.
	Selector string

	// Position is the character position where the error occurred.
	Position int

	// Message describes the error.
	Message string
}

// Error implements the error interface.
func (e *SelectorError) Error() string {
	return fmt.Sprintf("invalid selector %q at position %d: %s", e.Selector, e.Position, e.Message)
}
