package main

import "testing"

// It is a good idea to write a test to validate your code works at test/build time.
func TestFloBuilder(t *testing.T) {
	fb := FloBuilder()
	got := fb.Validate()
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
