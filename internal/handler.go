package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

func RecallsHandler(w http.ResponseWriter, r *http.Request) {
	page := getQueryInt(r, "page", 1)
	pageSize := getQueryInt(r, "pageSize", 20)

	category := r.URL.Query().Get("category")
	zone := r.URL.Query().Get("zone")
	brand := r.URL.Query().Get("brand")
	risk := r.URL.Query().Get("risk")
	dateStart := r.URL.Query().Get("dateStart")
	dateEnd := r.URL.Query().Get("dateEnd")

	recalls, err := GetPaginatedRecallsFiltered(page, pageSize, category, zone, risk, brand, dateStart, dateEnd)
	if err != nil {
		http.Error(w, "Failed to fetch recalls: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Attach links if you want to keep that behavior
	for i := range recalls {
		recalls[i].Links = []Link{
			{Rel: "self", Href: fmt.Sprintf("/recalls/%d", recalls[i].ID)},
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(recalls)
}

func RecallDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/recall/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid recall ID", http.StatusBadRequest)
		return
	}

	recall, err := GetRecallByID(id)
	if err != nil {
		http.Error(w, "recall not found", http.StatusNotFound)
		return
	}

	// Attach HATEOAS links directly
	recall.Links = []Link{
		{Rel: "self", Href: fmt.Sprintf("/recall/%d", recall.ID)},
		{Rel: "collection", Href: "/recalls"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recall) // ✅ just the Recall object
}

// Health check: just confirm server is alive
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Readiness check: confirm DB is connected
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if DB == nil {
		http.Error(w, "db not initialized", http.StatusServiceUnavailable)
		return
	}

	if err := DB.Ping(); err != nil {
		http.Error(w, "db not reachable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
func RecallsHTMLOrJSONHandler(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")

	if strings.Contains(accept, "application/json") {
		RecallsHandler(w, r)
		return
	}

	ListRecallsHandler(w, r)
}

func FetchAndUpsertHandler(w http.ResponseWriter, r *http.Request) {
	recalls, err := FetchRecalls()
	if err != nil {
		http.Error(w, "Failed to fetch from upstream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var success, failed int
	for _, recall := range recalls {
		if err := UpsertRecall(recall); err != nil {
			log.Println("Upsert error:", err)
			failed++
		} else {
			success++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"inserted_or_updated": success,
		"errors":              failed,
	})
}

func ListRecallsHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	category := r.URL.Query().Get("category")
	zone := r.URL.Query().Get("zone")
	brand := r.URL.Query().Get("brand")
	risk := r.URL.Query().Get("risk") // NEW: get risk filter
	dateStart := r.URL.Query().Get("dateStart")
	dateEnd := r.URL.Query().Get("dateEnd")

	recalls, err := GetPaginatedRecallsFiltered(page, 20, category, zone, risk, brand, dateStart, dateEnd)
	if err != nil {
		http.Error(w, "Error loading recalls", http.StatusInternalServerError)
		return
	}
	// for dropdowns
	categories, err := GetAllCategories()
	if err != nil {
		http.Error(w, "Failed to load categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	zones, err := GetAllZones()
	if err != nil {
		http.Error(w, "Failed to load zones: "+err.Error(), http.StatusInternalServerError)
		return
	}

	brands, err := GetAllBrands()
	if err != nil {
		http.Error(w, "Failed to load brands: "+err.Error(), http.StatusInternalServerError)
		return
	}

	risks, err := GetAllRisks()
	if err != nil {
		http.Error(w, "Failed to load risks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"minus": minus,
		"add":   add,
	}

	tmpl := template.Must(template.New("recalls.html").Funcs(funcMap).ParseFiles("templates/recalls.html"))
	data := RecallPageData{
		Recalls:          recalls,
		Categories:       categories,
		Zones:            zones,
		Brands:           brands,
		Risks:            risks,
		SelectedCategory: category,
		SelectedZone:     zone,
		SelectedBrand:    brand,
		SelectedRisk:     risk,
		DateStart:        dateStart,
		DateEnd:          dateEnd,
		Page:             page,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}

}

// tiny helper
func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	cats, err := GetAllCategories()
	if err != nil {
		http.Error(w, "failed to load categories: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  cats,
		"count": len(cats),
	}, http.StatusOK)
}

func RisksHandler(w http.ResponseWriter, r *http.Request) {
	risks, err := GetAllRisks()
	if err != nil {
		http.Error(w, "failed to load risks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  risks,
		"count": len(risks),
	}, http.StatusOK)
}

func ZonesHandler(w http.ResponseWriter, r *http.Request) {
	zones, err := GetAllZones()
	if err != nil {
		http.Error(w, "failed to load zones: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  zones,
		"count": len(zones),
	}, http.StatusOK)
}

func BrandsHandler(w http.ResponseWriter, r *http.Request) {
	brands, err := GetAllBrands()
	if err != nil {
		http.Error(w, "failed to load brands: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  brands,
		"count": len(brands),
	}, http.StatusOK)
}

// Optional: one-shot endpoint to fetch all filters in one request
func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	cats, _ := GetAllCategories()
	risks, _ := GetAllRisks()
	zones, _ := GetAllZones()
	brands, _ := GetAllBrands()

	writeJSON(w, map[string]any{
		"categories": cats,
		"risks":      risks,
		"zones":      zones,
		"brands":     brands,
	}, http.StatusOK)
}
