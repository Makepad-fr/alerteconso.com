package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Optionally, mock FetchRecalls for the handler test
// net/http/httptest is used to simulate real HTTP requests without a running server

func TestRecallsHandler_StatusOK(t *testing.T) {
	// ✅ Setup DB connection (same as your main.go)
	InitDB("postgres://rappeluser:rappelpass@localhost:5432/rappeldb?sslmode=disable")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/recalls", nil)

	// 🧪 This will call FetchRecalls() and UpsertRecall() internally
	RecallsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}
