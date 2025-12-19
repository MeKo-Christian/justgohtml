package errors_test

import (
	"errors"
	"strings"
	"testing"

	htmlerrors "github.com/MeKo-Christian/JustGoHTML/errors"
)

func TestParseError(t *testing.T) {
	t.Parallel()

	t.Run("Error with line and column", func(t *testing.T) {
		err := &htmlerrors.ParseError{
			Code:    "unexpected-null-character",
			Message: "Unexpected null character found",
			Line:    10,
			Column:  25,
		}

		expected := "unexpected-null-character at 10:25: Unexpected null character found"
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Error without location", func(t *testing.T) {
		err := &htmlerrors.ParseError{
			Code:    "eof-in-tag",
			Message: "Unexpected end of file in tag",
			Line:    0,
			Column:  0,
		}

		expected := "eof-in-tag: Unexpected end of file in tag"
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Error with only line", func(t *testing.T) {
		err := &htmlerrors.ParseError{
			Code:    "test-error",
			Message: "Test message",
			Line:    5,
			Column:  0,
		}

		// When column is 0, should not include location
		expected := "test-error: Test message"
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	t.Run("Empty errors", func(t *testing.T) {
		errs := htmlerrors.ParseErrors{}
		expected := "no parse errors"
		if got := errs.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Single error", func(t *testing.T) {
		errs := htmlerrors.ParseErrors{
			{
				Code:    "test-error",
				Message: "Test message",
				Line:    1,
				Column:  1,
			},
		}

		expected := "test-error at 1:1: Test message"
		if got := errs.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Multiple errors", func(t *testing.T) {
		errs := htmlerrors.ParseErrors{
			{
				Code:    "error-one",
				Message: "First error",
				Line:    1,
				Column:  10,
			},
			{
				Code:    "error-two",
				Message: "Second error",
				Line:    5,
				Column:  20,
			},
			{
				Code:    "error-three",
				Message: "Third error",
			},
		}

		result := errs.Error()

		// Check that it starts with the count
		if !strings.HasPrefix(result, "3 parse errors:\n") {
			t.Errorf("Error() should start with '3 parse errors:\\n', got %q", result)
		}

		// Check that all errors are included
		if !strings.Contains(result, "error-one at 1:10: First error") {
			t.Error("Error() should contain first error")
		}
		if !strings.Contains(result, "error-two at 5:20: Second error") {
			t.Error("Error() should contain second error")
		}
		if !strings.Contains(result, "error-three: Third error") {
			t.Error("Error() should contain third error")
		}

		// Check formatting with newlines
		if !strings.Contains(result, "\n  - ") {
			t.Error("Error() should have proper formatting with newlines and bullets")
		}
	})

	t.Run("Unwrap returns error slice", func(t *testing.T) {
		err1 := &htmlerrors.ParseError{Code: "err1", Message: "Error 1"}
		err2 := &htmlerrors.ParseError{Code: "err2", Message: "Error 2"}
		errs := htmlerrors.ParseErrors{err1, err2}

		unwrapped := errs.Unwrap()
		if len(unwrapped) != 2 {
			t.Errorf("Unwrap() returned %d errors, want 2", len(unwrapped))
		}

		// Verify the errors are the same
		if !errors.Is(unwrapped[0], err1) {
			t.Error("Unwrap()[0] should be err1")
		}
		if !errors.Is(unwrapped[1], err2) {
			t.Error("Unwrap()[1] should be err2")
		}
	})

	t.Run("Unwrap with empty errors", func(t *testing.T) {
		errs := htmlerrors.ParseErrors{}
		unwrapped := errs.Unwrap()
		if len(unwrapped) != 0 {
			t.Errorf("Unwrap() returned %d errors, want 0", len(unwrapped))
		}
	})
}

func TestSelectorError(t *testing.T) {
	t.Parallel()

	t.Run("Error with all fields", func(t *testing.T) {
		err := &htmlerrors.SelectorError{
			Selector: "div > .class[invalid",
			Position: 15,
			Message:  "unclosed attribute selector",
		}

		expected := `invalid selector "div > .class[invalid" at position 15: unclosed attribute selector`
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Error at position 0", func(t *testing.T) {
		err := &htmlerrors.SelectorError{
			Selector: "*invalid",
			Position: 0,
			Message:  "unexpected character at start",
		}

		expected := `invalid selector "*invalid" at position 0: unexpected character at start`
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})
}

func TestErrNotImplemented(t *testing.T) {
	t.Parallel()

	if htmlerrors.ErrNotImplemented == nil {
		t.Fatal("ErrNotImplemented should not be nil")
	}

	expected := "not implemented"
	if got := htmlerrors.ErrNotImplemented.Error(); got != expected {
		t.Errorf("ErrNotImplemented.Error() = %q, want %q", got, expected)
	}

	// Verify it can be used with errors.Is
	if !errors.Is(htmlerrors.ErrNotImplemented, htmlerrors.ErrNotImplemented) {
		t.Error("errors.Is should work with ErrNotImplemented")
	}
}

func TestMessage(t *testing.T) {
	t.Parallel()

	t.Run("Known error code", func(t *testing.T) {
		// Test with a known error code from codes.go
		msg := htmlerrors.Message("eof-in-tag")
		if msg == "" {
			t.Error("Message() should return a non-empty string for known error code")
		}
		if msg == "Unknown error" {
			t.Error("Message() should not return 'Unknown error' for known error code")
		}
	})

	t.Run("Unknown error code", func(t *testing.T) {
		msg := htmlerrors.Message("this-error-does-not-exist")
		expected := "Unknown error"
		if msg != expected {
			t.Errorf("Message() = %q, want %q for unknown error code", msg, expected)
		}
	})

	t.Run("Empty error code", func(t *testing.T) {
		msg := htmlerrors.Message("")
		expected := "Unknown error"
		if msg != expected {
			t.Errorf("Message() = %q, want %q for empty error code", msg, expected)
		}
	})
}
