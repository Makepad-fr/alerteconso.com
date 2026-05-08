package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Optionally, mock FetchRecalls for the handler test
// net/http/httptest is used to simulate real HTTP requests without a running server

func TestRecallCollectionHandlerStatusOK(t *testing.T) {
	// ✅ Setup DB connection (same as your main.go)
	InitDB(testDatabaseURL(t))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/recalls", nil)

	RecallCollectionHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func mapQuery(key, value string) url.Values {
	values := url.Values{}
	values.Set(key, value)
	return values
}

func TestCollectionLinks(t *testing.T) {
	query := mapQuery("q", "fromage")
	links := collectionLinks("/recalls", query, 2, 20, true)

	want := map[string]string{
		"self":  "/recalls?page=2&pageSize=20&q=fromage",
		"first": "/recalls?page=1&pageSize=20&q=fromage",
		"prev":  "/recalls?page=1&pageSize=20&q=fromage",
		"next":  "/recalls?page=3&pageSize=20&q=fromage",
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

func TestCollectionLinksOmitsNextWhenThereIsNoNextPage(t *testing.T) {
	links := collectionLinks("/recalls", url.Values{}, 1, 20, false)
	for _, link := range links {
		if link.Rel == "next" {
			t.Fatalf("expected no next link, got %q", link.Href)
		}
	}
}

func TestRecallCollectionHandlerRejectsOversizedPageSize(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/recalls?pageSize=101", nil)

	RecallCollectionHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
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
		"self":       "/recalls/123",
		"collection": "/recalls",
		"official":   "https://rappel.conso.gouv.fr/fiche-rappel/123/interne",
		"pdf":        "https://rappel.conso.gouv.fr/affichettepdf/123/interne",
	} {
		if got[rel] != href {
			t.Fatalf("expected %s link %q, got %q", rel, href, got[rel])
		}
	}
}

func TestRecallSummariesAttachLinks(t *testing.T) {
	summaries := recallSummaries([]Recall{{
		ID:                    123,
		NumeroFiche:           "2026-05-0001",
		LienVersLaFicheRappel: "https://rappel.conso.gouv.fr/fiche-rappel/123/interne",
		LienVersAffichettePDF: "https://rappel.conso.gouv.fr/affichettepdf/123/interne",
	}})

	if len(summaries) != 1 {
		t.Fatalf("expected one summary, got %d", len(summaries))
	}
	got := map[string]string{}
	for _, link := range summaries[0].Links {
		got[link.Rel] = link.Href
	}
	if got["self"] != "/recalls/123" || got["collection"] != "/recalls" || got["official"] == "" || got["pdf"] == "" {
		t.Fatalf("unexpected summary links: %#v", got)
	}
}

func TestRecallIDFromPathOnlyAcceptsCanonicalAPIPath(t *testing.T) {
	if got := recallIDFromPath("/recalls/123"); got != "123" {
		t.Fatalf("expected canonical recall id 123, got %q", got)
	}
	if got := recallIDFromPath("/recall/123"); got != "" {
		t.Fatalf("expected noncanonical recall path to be ignored, got %q", got)
	}
	if got := recallIDFromPath("/api/recalls/123"); got != "" {
		t.Fatalf("expected namespaced API path to be ignored, got %q", got)
	}
}

func TestCanonicalRecallPath(t *testing.T) {
	for path, expected := range map[string]string{
		"/recalls/":            "/recalls",
		"/recalls/123/":        "/recalls/123",
		"/recalls/categories/": "/recalls/categories",
	} {
		got, ok := canonicalRecallPath(path)
		if !ok || got != expected {
			t.Fatalf("expected %q to canonicalize to %q, got %q with ok=%v", path, expected, got, ok)
		}
	}

	if got, ok := canonicalRecallPath("/recalls/123"); ok {
		t.Fatalf("expected canonical path to be unchanged, got %q", got)
	}
}

func TestRequireMethodAllowsHeadForGetResources(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/recalls", nil)
	if !requireMethod(rr, req, http.MethodGet) {
		t.Fatal("expected HEAD to be allowed for GET resource")
	}
}

func TestValidateDateFilters(t *testing.T) {
	if err := validateDateFilters("2026-05-01", "2026-05-08T00:00:00.123Z"); err != nil {
		t.Fatalf("expected valid date filters, got %v", err)
	}
	if err := validateDateFilters("not-a-date", ""); err == nil {
		t.Fatal("expected invalid start date to fail")
	}
	if err := validateDateFilters("2026-05-09", "2026-05-08"); err == nil {
		t.Fatal("expected inverted date range to fail")
	}
}

func TestPrefersHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if !PrefersHTML(req) {
		t.Fatal("expected browser Accept header to prefer HTML")
	}

	req = httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "application/json,text/html;q=0.9")
	if PrefersHTML(req) {
		t.Fatal("expected application/json to prefer JSON")
	}

	req = httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "*/*")
	if PrefersHTML(req) {
		t.Fatal("expected wildcard Accept header to use JSON representation")
	}

	req = httptest.NewRequest("GET", "/recalls", nil)
	req.Header.Set("Accept", "text/html;q=2,application/json")
	if PrefersHTML(req) {
		t.Fatal("expected invalid HTML q value to be ignored")
	}
}
