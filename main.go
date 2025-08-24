package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Makepad-fr/rappelconsommation/internal"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}
	internal.InitDB(dbURL)

	// ✅ Defer DB close
	defer func() {
		if internal.DB != nil {
			internal.DB.Close()
			fmt.Println("🔌 PostgreSQL connection closed")
		}
	}()
	// Start background cron job every 5 min
	go startRecallUpdater()

	http.HandleFunc("/recall/", internal.RecallDetailHandler)
	http.HandleFunc("/", internal.ListRecallsHandler)
	http.HandleFunc("/recalls", internal.RecallsHTMLOrJSONHandler)
	http.HandleFunc("/healthz", internal.HealthzHandler)
	http.HandleFunc("/readyz", internal.ReadyzHandler)
	http.HandleFunc("/categories", internal.CategoriesHandler)
	http.HandleFunc("/risks", internal.RisksHandler)
	http.HandleFunc("/zones", internal.ZonesHandler) // or /locations
	http.HandleFunc("/brands", internal.BrandsHandler)

	// optional all-in-one
	http.HandleFunc("/filters", internal.FiltersHandler)
	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func startRecallUpdater() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately once on startup
	upsertRecalls()

	for range ticker.C {
		upsertRecalls()
	}
}

func upsertRecalls() {
	fmt.Println("🔁 Fetching recalls from data.gouv.fr...")

	recalls, err := internal.FetchRecalls()
	if err != nil {
		log.Println("❌ Failed to fetch recalls:", err)
		return
	}

	count := 0
	for _, r := range recalls {
		if err := internal.UpsertRecall(r); err != nil {
			log.Printf("❌ Failed to upsert recall %d: %v\n", r.ID, err)
			continue
		}
		count++
	}

	fmt.Printf("✅ Upserted %d recalls at %s\n", count, time.Now().Format(time.RFC822))
}
