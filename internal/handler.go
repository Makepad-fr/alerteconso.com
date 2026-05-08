package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

func APIRootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api" && r.URL.Path != "/api/" {
		http.NotFound(w, r)
		return
	}
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	writeJSON(w, map[string]any{
		"_links": []Link{
			{Rel: "self", Href: "/api"},
			{Rel: "recalls", Href: "/api/recalls"},
			{Rel: "recall-filters", Href: "/api/recalls/filters"},
			{Rel: "recall-categories", Href: "/api/recalls/categories"},
			{Rel: "recall-risks", Href: "/api/recalls/risks"},
			{Rel: "recall-zones", Href: "/api/recalls/zones"},
			{Rel: "recall-brands", Href: "/api/recalls/brands"},
		},
	}, http.StatusOK)
}

func APIRecallsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/recalls" {
		http.NotFound(w, r)
		return
	}
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	page := getQueryInt(r, "page", 1)
	pageSize := getQueryInt(r, "pageSize", 20)

	recalls, err := getRecallsFromRequest(r, page, pageSize)
	if err != nil {
		http.Error(w, "Failed to fetch recalls: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range recalls {
		attachRecallLinks(&recalls[i])
	}

	links := collectionLinks("/api/recalls", r.URL.Query(), page, pageSize, len(recalls))
	writeCollectionLinkHeader(w, "/api/recalls", r.URL.Query(), page, pageSize, len(recalls))
	writeJSON(w, RecallListResponse{
		Data: recalls,
		Page: PageMeta{
			Page:     page,
			PageSize: pageSize,
			Count:    len(recalls),
		},
		Links: links,
	}, http.StatusOK)
}

func RecallDetailHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	idStr := recallIDFromPath(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recall, err := GetRecallByID(id)
	if err != nil {
		http.Error(w, "recall not found", http.StatusNotFound)
		return
	}

	// Attach HATEOAS links directly
	attachRecallLinks(&recall)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(recall)
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
func RecallsPageHandler(w http.ResponseWriter, r *http.Request) {
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := r.URL.Query().Get("category")
	zone := r.URL.Query().Get("zone")
	brand := r.URL.Query().Get("brand")
	risk := r.URL.Query().Get("risk")
	dateStart := r.URL.Query().Get("dateStart")
	dateEnd := r.URL.Query().Get("dateEnd")

	var (
		recalls []Recall
		err     error
	)
	if q != "" {
		recalls, err = SearchRecalls(page, 20, q)
	} else {
		recalls, err = GetPaginatedRecallsFiltered(page, 20, category, zone, risk, brand, dateStart, dateEnd)
	}
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
		Query:            q,
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	cats, err := GetAllCategories()
	if err != nil {
		http.Error(w, "failed to load categories: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  cats,
		"count": len(cats),
		"_links": []Link{
			{Rel: "self", Href: r.URL.Path},
			{Rel: "recalls", Href: "/api/recalls"},
		},
	}, http.StatusOK)
}

func RisksHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	risks, err := GetAllRisks()
	if err != nil {
		http.Error(w, "failed to load risks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  risks,
		"count": len(risks),
		"_links": []Link{
			{Rel: "self", Href: r.URL.Path},
			{Rel: "recalls", Href: "/api/recalls"},
		},
	}, http.StatusOK)
}

func ZonesHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	zones, err := GetAllZones()
	if err != nil {
		http.Error(w, "failed to load zones: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  zones,
		"count": len(zones),
		"_links": []Link{
			{Rel: "self", Href: r.URL.Path},
			{Rel: "recalls", Href: "/api/recalls"},
		},
	}, http.StatusOK)
}

func BrandsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	brands, err := GetAllBrands()
	if err != nil {
		http.Error(w, "failed to load brands: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"data":  brands,
		"count": len(brands),
		"_links": []Link{
			{Rel: "self", Href: r.URL.Path},
			{Rel: "recalls", Href: "/api/recalls"},
		},
	}, http.StatusOK)
}

// Optional: one-shot endpoint to fetch all filters in one request
func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	cats, _ := GetAllCategories()
	risks, _ := GetAllRisks()
	zones, _ := GetAllZones()
	brands, _ := GetAllBrands()

	writeJSON(w, map[string]any{
		"categories": cats,
		"risks":      risks,
		"zones":      zones,
		"brands":     brands,
		"_links": []Link{
			{Rel: "self", Href: r.URL.Path},
			{Rel: "recalls", Href: "/api/recalls"},
			{Rel: "recall-categories", Href: "/api/recalls/categories"},
			{Rel: "recall-risks", Href: "/api/recalls/risks"},
			{Rel: "recall-zones", Href: "/api/recalls/zones"},
			{Rel: "recall-brands", Href: "/api/recalls/brands"},
		},
	}, http.StatusOK)
}

func getRecallsFromRequest(r *http.Request, page, pageSize int) ([]Recall, error) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := r.URL.Query().Get("category")
	zone := r.URL.Query().Get("zone")
	brand := r.URL.Query().Get("brand")
	risk := r.URL.Query().Get("risk")
	dateStart := r.URL.Query().Get("dateStart")
	dateEnd := r.URL.Query().Get("dateEnd")

	if q != "" {
		return SearchRecalls(page, pageSize, q)
	}
	return GetPaginatedRecallsFiltered(page, pageSize, category, zone, risk, brand, dateStart, dateEnd)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func recallIDFromPath(path string) string {
	const prefix = "/api/recalls/"
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return ""
}

func attachRecallLinks(recall *Recall) {
	links := FlexibleLinks{
		{Rel: "self", Href: fmt.Sprintf("/api/recalls/%d", recall.ID)},
		{Rel: "collection", Href: "/api/recalls"},
	}
	if recall.LienVersLaFicheRappel != "" {
		links = append(links, Link{Rel: "official", Href: recall.LienVersLaFicheRappel})
	}
	if recall.LienVersAffichettePDF != "" {
		links = append(links, Link{Rel: "pdf", Href: recall.LienVersAffichettePDF})
	}
	recall.Links = links
}

func collectionLinks(path string, query url.Values, page, pageSize, count int) []Link {
	links := []Link{
		{Rel: "self", Href: paginatedURL(path, query, page, pageSize)},
		{Rel: "first", Href: paginatedURL(path, query, 1, pageSize)},
	}
	if page > 1 {
		links = append(links, Link{Rel: "prev", Href: paginatedURL(path, query, page-1, pageSize)})
	}
	if count >= pageSize {
		links = append(links, Link{Rel: "next", Href: paginatedURL(path, query, page+1, pageSize)})
	}
	return links
}

func writeCollectionLinkHeader(w http.ResponseWriter, path string, query url.Values, page, pageSize, count int) {
	parts := make([]string, 0)
	for _, link := range collectionLinks(path, query, page, pageSize, count) {
		parts = append(parts, fmt.Sprintf("<%s>; rel=%q", link.Href, link.Rel))
	}
	w.Header().Set("Link", strings.Join(parts, ", "))
}

func paginatedURL(path string, query url.Values, page, pageSize int) string {
	values := url.Values{}
	for key, vals := range query {
		for _, value := range vals {
			values.Add(key, value)
		}
	}
	values.Set("page", strconv.Itoa(page))
	values.Set("pageSize", strconv.Itoa(pageSize))
	if encoded := values.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
