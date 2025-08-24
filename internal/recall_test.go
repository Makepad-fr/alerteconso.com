package internal

import (
	"encoding/json"
	"os"
	"testing"
)

//Test... prefix = Go recognizes this function as a test

func TestRecallUnmarshal(t *testing.T) {
	// Read test file
	data, err := os.ReadFile("../testdata/sample.json")
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}

	var recalls []Recall
	if err := json.Unmarshal(data, &recalls); err != nil {
		t.Fatalf("failed to unmarshal sample json: %v", err)
	}

	if len(recalls) != 1 {
		t.Errorf("expected 1 recall, got %d", len(recalls))
	}

	recall := recalls[0]
	if recall.ID != 1 {
		t.Errorf("expected ID 1, got %d", recall.ID)
	}
	if recall.MarqueProduit != "TestBrand" {
		t.Errorf("expected MarqueProduit 'TestBrand', got %q", recall.MarqueProduit)
	}
	if recall.Libelle != "Test Recall" {
		t.Errorf("expected Libelle 'Test Recall', got %q", recall.Libelle)
	}
}
