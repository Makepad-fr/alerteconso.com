package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Makepad-fr/rappelconsommation/internal"
)

//go:embed index.html
var indexHTML []byte

//go:embed logo.png
var logoPNG []byte

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

	// Pick listen port from env (default 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/recalls/categories", internal.CategoriesHandler)
	http.HandleFunc("/recalls/risks", internal.RisksHandler)
	http.HandleFunc("/recalls/zones", internal.ZonesHandler)
	http.HandleFunc("/recalls/brands", internal.BrandsHandler)
	http.HandleFunc("/recalls/filters", internal.FiltersHandler)
	http.HandleFunc("/recalls/", internal.RecallDetailHandler)
	http.HandleFunc("/recalls", internal.RecallsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !internal.PrefersHTML(r) {
			internal.RootHandler(w, r)
			return
		}
		// If filters/search are submitted to '/', redirect to '/recalls' keeping the query string
		if r.URL.RawQuery != "" {
			http.Redirect(w, r, "/recalls?"+r.URL.RawQuery, http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	// serve embedded logo (referenced as href="/logo.png" or "logo.png")
	http.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(logoPNG)
	})
	http.HandleFunc("/healthz", internal.HealthzHandler)
	http.HandleFunc("/readyz", internal.ReadyzHandler)
	fmt.Printf("Listening on http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
