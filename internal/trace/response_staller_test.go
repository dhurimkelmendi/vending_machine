package trace_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhurimkelmendi/vending_machine/internal/trace"
)

func TestResponseStaller(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	st := trace.NewResponseStaller(w)
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	expectedStatus := http.StatusNoContent
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatus)
	}
	handler(st, r)

	if st.Status != expectedStatus {
		t.Fatalf("expected status %d but got %d", expectedStatus, st.Status)
	}

	if w.Result().StatusCode != expectedStatus {
		t.Fatalf("expected status %d but got %d", expectedStatus, w.Result().StatusCode)
	}
}
