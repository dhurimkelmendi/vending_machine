package trace_test

import (
	"strings"
	"testing"

	"github.com/dhurimkelmendi/vending_machine/internal/trace"
)

func TestTrace(t *testing.T) {
	t.Parallel()

	s := trace.Getfl()
	if !strings.Contains(s, "13") {
		t.Fatalf("expected trace to contain 13 but got %s", s)
	}
	if !strings.Contains(s, "TestTrace") {
		t.Fatalf("expected trace to contain TestTrace but got %s", s)
	}
}
