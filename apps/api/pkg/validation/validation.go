// Package validation provides the small set of request checks the MVP needs.
//
// Validation is explicit and per-endpoint rather than reflection-driven, so the
// rules stay readable and the messages stay user-facing.
package validation

import (
	"strings"
	"unicode/utf8"
)

// Errors accumulates field-level failures keyed by request field name.
type Errors map[string]string

// Add records a failure for a field, keeping the first message per field.
func (e *Errors) Add(field, message string) {
	if *e == nil {
		*e = make(Errors)
	}
	if _, exists := (*e)[field]; !exists {
		(*e)[field] = message
	}
}

// Any reports whether any failure was recorded.
func (e Errors) Any() bool { return len(e) > 0 }

// Required records a failure when value is blank.
func (e *Errors) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		e.Add(field, "This field is required.")
	}
}

// MaxLength records a failure when value exceeds max runes.
func (e *Errors) MaxLength(field, value string, max int) {
	if utf8.RuneCountInString(value) > max {
		e.Add(field, "This value is too long.")
	}
}

// TrimmedLength returns the rune count of value with surrounding space removed.
func TrimmedLength(value string) int {
	return utf8.RuneCountInString(strings.TrimSpace(value))
}
