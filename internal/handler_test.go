package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Optionally, mock FetchRecalls for the handler test
// net/http/httptest is used to simulate real HTTP requests without a running server

func TestRecallsHandler_StatusOK(t *testing.T) {
	// ✅ Setup DB connection (same as your main.go)
	InitDB(testDatabaseURL(t))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/recalls", nil)

	// 🧪 This will call FetchRecalls() and UpsertRecall() internally
	RecallsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func mapQuery(key, value string) url.Values {
	values := url.Values{}
	values.Set(key, value)
	return values
}

func TestWantsJSONKeepsLegacyAPIDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/recalls", nil)
	if !wantsJSON(req) {
		t.Fatal("expected missing Accept header to return legacy JSON API")
	}

	req = httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "*/*")
	if !wantsJSON(req) {
		t.Fatal("expected */* Accept header to return legacy JSON API")
	}

	req = httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	if wantsJSON(req) {
		t.Fatal("expected browser HTML Accept header to render HTML page")
	}
}

func TestCollectionLinks(t *testing.T) {
	query := mapQuery("q", "fromage")
	links := collectionLinks("/api/recalls", query, 2, 20, 20)

	want := map[string]string{
		"self":  "/api/recalls?page=2&pageSize=20&q=fromage",
		"first": "/api/recalls?page=1&pageSize=20&q=fromage",
		"prev":  "/api/recalls?page=1&pageSize=20&q=fromage",
		"next":  "/api/recalls?page=3&pageSize=20&q=fromage",
	}
	for _, link := range links {
		if expected, ok := want[link.Rel]; ok && link.Href != expected {
			t.Fatalf("expected %s link %q, got %q", link.Rel, expected, link.Href)
		}
		delete(want, link.Rel)
	}
	if len(want) > 0 {
		t.Fatalf("missing links: %#v", want)
	}
}

func TestAttachRecallLinks(t *testing.T) {
	recall := Recall{
		ID:                       123,
		LienVersLaFicheRappel:    "https://rappel.conso.gouv.fr/fiche-rappel/123/interne",
		LienVersAffichettePDF:    "https://rappel.conso.gouv.fr/affichettepdf/123/interne",
		IdentificationProduits:   "lot-a",
		LiensVersLesImagesRaw:    "",
		DatePublication:          "2026-05-08T00:00:00Z",
		NumeroFiche:              "2026-05-0001",
		CategorieProduit:         "alimentation",
		SousCategorieProduit:     "fromage",
		MarqueProduit:            "test",
		ModelesOuReferences:      "ref",
		RisquesEncourus:          "listeria",
		MotifRappel:              "test",
		PreconisationsSanitaires: "test",
	}
	attachRecallLinks(&recall)

	got := map[string]string{}
	for _, link := range recall.Links {
		got[link.Rel] = link.Href
	}
	for rel, href := range map[string]string{
		"self":           "/recalls/123",
		"api":            "/api/recalls/123",
		"collection":     "/recalls",
		"api-collection": "/api/recalls",
		"official":       "https://rappel.conso.gouv.fr/fiche-rappel/123/interne",
		"pdf":            "https://rappel.conso.gouv.fr/affichettepdf/123/interne",
	} {
		if got[rel] != href {
			t.Fatalf("expected %s link %q, got %q", rel, href, got[rel])
		}
	}
}
