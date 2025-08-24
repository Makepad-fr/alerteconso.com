package internal

import (
	"testing"
)

// ⚠️ Make sure InitDB is called in main_test.go or before running this test manually

func TestFetchAndUpsertRecalls(t *testing.T) {
	InitDB("postgres://rappeluser:rappelpass@localhost:5432/rappeldb?sslmode=disable")

	recalls, err := FetchRecalls()
	if err != nil {
		t.Fatalf("failed to fetch recalls: %v", err)
	}

	if len(recalls) == 0 {
		t.Fatal("expected some recalls, got 0")
	}

	for _, r := range recalls[:10] { // limit to 10 for testing
		if err := UpsertRecall(r); err != nil {
			t.Errorf("failed to upsert recall ID %d: %v", r.ID, err)
		}
	}
}
