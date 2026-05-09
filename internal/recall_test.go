package internal

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
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

func TestRecallUnmarshalV2IdentificationArray(t *testing.T) {
	data := []byte(`[
		{
			"id": 49788,
			"rappel_guid": "59671e743032024fdf5f4cadc6865df1a8371a0c",
			"numero_fiche": "sr/01361/26",
			"numero_version": 0,
			"categorie_produit": "vêtements, mode, epi",
			"sous_categorie_produit": "vêtements, textiles, accessoires de mode",
			"marque_produit": "xmgolong",
			"modeles_ou_references": "chaussures de sport pour enfants",
			"identification_produits": ["gtin", "lot", "2026-05-08"],
			"risques_encourus": "environnement|risque chimique",
			"lien_vers_affichette_pdf": "https://rappel.conso.gouv.fr/affichettepdf/49788/rapex",
			"lien_vers_la_fiche_rappel": "https://rappel.conso.gouv.fr/fiche-rappel/49788/rapex",
			"date_publication": "2026-05-08T00:00:00+00:00",
			"libelle": "chaussures pour enfants",
			"_links": {}
		}
	]`)

	var recalls []Recall
	if err := json.Unmarshal(data, &recalls); err != nil {
		t.Fatalf("failed to unmarshal v2 json: %v", err)
	}
	if got, want := recalls[0].IdentificationProduits.String(), "gtin | lot | 2026-05-08"; got != want {
		t.Fatalf("expected identification products %q, got %q", want, got)
	}
}

func TestFlexibleLinksRejectsUnsupportedShape(t *testing.T) {
	var links FlexibleLinks
	if err := json.Unmarshal([]byte(`{"unexpected":"shape"}`), &links); err == nil {
		t.Fatal("expected unsupported non-empty link object to fail")
	}
}

func TestFlexibleLinksAllowsEmptyObject(t *testing.T) {
	var links FlexibleLinks
	if err := json.Unmarshal([]byte(`{}`), &links); err != nil {
		t.Fatalf("expected empty object to be accepted, got %v", err)
	}
	if links != nil {
		t.Fatalf("expected empty object to produce nil links, got %#v", links)
	}
}

func TestNormalizeDatePublicationForDBConvertsOffsetsToUTC(t *testing.T) {
	got, err := normalizeDatePublicationForDB("2026-05-08T02:30:00+02:00")
	if err != nil {
		t.Fatalf("expected valid date publication, got %v", err)
	}
	if !got.Valid {
		t.Fatal("expected normalized date publication to be valid")
	}
	if formatted := got.Time.Format(time.RFC3339); formatted != "2026-05-08T00:30:00Z" {
		t.Fatalf("expected UTC-normalized date publication, got %q", formatted)
	}
}

func TestNormalizeDatePublicationForDBRejectsInvalidDate(t *testing.T) {
	if _, err := normalizeDatePublicationForDB("not-a-date"); err == nil {
		t.Fatal("expected invalid date publication to fail")
	}
}

func TestNormalizeDatePublicationForDBRejectsMissingDate(t *testing.T) {
	if _, err := normalizeDatePublicationForDB(""); err == nil {
		t.Fatal("expected missing date publication to fail")
	}
}

func TestBuildRecallsCollectionQueryCombinesSearchAndFilters(t *testing.T) {
	start := dateFilterBound{Time: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Valid: true}
	end := dateFilterBound{Time: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC), Valid: true, Exclusive: true}

	query, args := buildRecallsCollectionQuery(
		21,
		40,
		" fromage ",
		"alimentation",
		"France",
		"listeria",
		"Brand",
		start,
		end,
	)

	for _, fragment := range []string{
		"libelle ILIKE $1",
		"categorie_produit = $2",
		"zone_geographique_de_vente = $3",
		"risques_encourus || '|') ILIKE '%' || '|' || $4 || '|' || '%'",
		"marque_produit = $5",
		"recalls.date_publication >= $6",
		"recalls.date_publication < $7",
		"LIMIT $8 OFFSET $9",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("expected query to contain %q, got %s", fragment, query)
		}
	}

	if len(args) != 9 {
		t.Fatalf("expected 9 query args, got %d: %#v", len(args), args)
	}
	if args[0] != "%fromage%" || args[1] != "alimentation" || args[2] != "France" || args[3] != "listeria" || args[4] != "Brand" {
		t.Fatalf("unexpected filter args: %#v", args[:5])
	}
	if args[5] != start.Time || args[6] != end.Time || args[7] != 21 || args[8] != 40 {
		t.Fatalf("unexpected pagination/date args: %#v", args[5:])
	}
}
