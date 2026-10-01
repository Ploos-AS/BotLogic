package logic

import "fmt"

// Validate compiles Prolog source in a fresh interpreter without mutating a
// Store, creating files, or changing a ruleset revision.
func Validate(source string) error {
	if source == "" { return fmt.Errorf("source is required") }
	e := New()
	return e.Consult(source)
}
