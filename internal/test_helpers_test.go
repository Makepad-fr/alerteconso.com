package internal

import (
	"os"
	"testing"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("set TEST_DATABASE_URL to run database-backed tests")
	}
	return dbURL
}
