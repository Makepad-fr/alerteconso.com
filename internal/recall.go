package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TODO: tum db querylerini logla ve ne kadar surdugunu logla
// fmt yerine log. kullan loglamal icin
const RappelURL = "https://www.data.gouv.fr/api/1/datasets/r/7b212733-7f5b-4ff3-b5b2-c7fea20f9cb1"

var rappelHTTPClient = &http.Client{Timeout: 45 * time.Second}

type RecallPageData struct {
	Recalls          []Recall
	Categories       []string
	Zones            []string
	Brands           []string
	Risks            []string
	SelectedCategory string
	SelectedZone     string
	SelectedBrand    string
	SelectedRisk     string
	DateStart        string
	DateEnd          string
	Query            string
	Page             int
}

type RecallResponse struct {
	Recall Recall `json:"recall"`
	Links  []Link `json:"_links"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type PageMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Count    int `json:"count"`
}

type RecallListResponse struct {
	Data  []Recall `json:"data"`
	Page  PageMeta `json:"page"`
	Links []Link   `json:"_links"`
}

type Recall struct {
	ID                                   int          `json:"id"`
	RappelGUID                           string       `json:"rappel_guid"`
	NumeroFiche                          string       `json:"numero_fiche"`
	NumeroVersion                        int          `json:"numero_version"`
	NatureJuridiqueRappel                string       `json:"nature_juridique_rappel"`
	CategorieProduit                     string       `json:"categorie_produit"`
	SousCategorieProduit                 string       `json:"sous_categorie_produit"`
	MarqueProduit                        string       `json:"marque_produit"`
	ModelesOuReferences                  string       `json:"modeles_ou_references"`
	IdentificationProduits               FlexibleText `json:"identification_produits"`
	Conditionnements                     string       `json:"conditionnements"`
	DateDebutCommercialisation           string       `json:"date_debut_commercialisation"`
	DateFinCommercialisation             string       `json:"date_date_fin_commercialisation"`
	DateLimiteDeConsommation             *string      `json:"date_limite_de_consommation"`
	TemperatureConservation              string       `json:"temperature_conservation"`
	MarqueSalubrite                      *string      `json:"marque_salubrite"`
	InformationsComplementaires          string       `json:"informations_complementaires"`
	ZoneGeographiqueDeVente              string       `json:"zone_geographique_de_vente"`
	Distributeurs                        string       `json:"distributeurs"`
	MotifRappel                          string       `json:"motif_rappel"`
	RisquesEncourus                      string       `json:"risques_encourus"`
	PreconisationsSanitaires             string       `json:"preconisations_sanitaires"`
	DescriptionComplementaireRisque      string       `json:"description_complementaire_risque"`
	ConduitesATenirParLeConsommateur     string       `json:"conduites_a_tenir_par_le_consommateur"`
	NumeroContact                        string       `json:"numero_contact"`
	ModalitesDeCompensation              string       `json:"modalites_de_compensation"`
	DateDeFinDeLaProcedureDeRappel       string       `json:"date_de_fin_de_la_procedure_de_rappel"`
	InformationsComplementairesPubliques string       `json:"informations_complementaires_publiques"`
	LiensVersLesImagesRaw                string       `json:"liens_vers_les_images"` // raw JSON string from API / DB
	ImageURLs                            []string     `json:"-"`                     // parsed URLs

	LienVersLaListeDesProduits      *string       `json:"lien_vers_la_liste_des_produits"`
	LienVersLaListeDesDistributeurs *string       `json:"lien_vers_la_liste_des_distributeurs"`
	LienVersAffichettePDF           string        `json:"lien_vers_affichette_pdf"`
	LienVersLaFicheRappel           string        `json:"lien_vers_la_fiche_rappel"`
	DatePublication                 string        `json:"date_publication"`
	Libelle                         string        `json:"libelle"`
	Links                           FlexibleLinks `json:"_links"`
}

func FetchRecalls() ([]Recall, error) {
	req, err := http.NewRequest(http.MethodGet, RappelURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "alerteconso.com/1.0")

	resp, err := rappelHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream returned %s", resp.Status)
	}

	var recalls []Recall
	if err := json.NewDecoder(resp.Body).Decode(&recalls); err != nil {
		return nil, err
	}

	for i := range recalls {
		recalls[i].ImageURLs = parseImageURLs(recalls[i].LiensVersLesImagesRaw)
	}

	return recalls, nil
}

// SearchRecalls provides a single free-text entry point over common fields.
// It does not alter existing handlers; you can call it when a `q` param is present.
func SearchRecalls(page, pageSize int, q string) ([]Recall, error) {
	return SearchRecallsWithLimit(page, pageSize, pageSize, q)
}

func SearchRecallsWithLimit(page, pageSize, limit int, q string) ([]Recall, error) {
	offset := (page - 1) * pageSize
	q = strings.TrimSpace(q)
	if q == "" {
		return GetPaginatedRecallsWithLimit(page, pageSize, limit)
	}

	term := "%" + q + "%"
	rows, err := DB.Query(`
		SELECT
			id, numero_fiche, categorie_produit, sous_categorie_produit,
			marque_produit, risques_encourus, motif_rappel,
			preconisations_sanitaires, numero_contact, distributeurs,
			modalites_de_compensation, zone_geographique_de_vente,
			lien_vers_affichette_pdf, lien_vers_la_fiche_rappel,
			liens_vers_les_images, libelle, date_publication
		FROM recalls
		WHERE (
			libelle ILIKE $1 OR
			marque_produit ILIKE $1 OR
			modeles_ou_references ILIKE $1 OR
			identification_produits ILIKE $1 OR
			numero_fiche ILIKE $1
		)
		ORDER BY date_publication DESC
		LIMIT $2 OFFSET $3
	`, term, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recalls []Recall
	for rows.Next() {
		var r Recall
		if err := rows.Scan(
			&r.ID, &r.NumeroFiche, &r.CategorieProduit, &r.SousCategorieProduit,
			&r.MarqueProduit, &r.RisquesEncourus, &r.MotifRappel,
			&r.PreconisationsSanitaires, &r.NumeroContact, &r.Distributeurs,
			&r.ModalitesDeCompensation, &r.ZoneGeographiqueDeVente,
			&r.LienVersAffichettePDF, &r.LienVersLaFicheRappel,
			&r.LiensVersLesImagesRaw, &r.Libelle, &r.DatePublication,
		); err != nil {
			return nil, err
		}
		r.ImageURLs = parseImageURLs(r.LiensVersLesImagesRaw)
		recalls = append(recalls, r)
	}
	return recalls, nil
}

// for each recall, inserts into the recalls table
// ON CONFLICT(id) to avoid duplication =  if a row with the same ID exists, it updates it
// Ensures that your DB is always up-to-date without duplicates

func UpsertRecall(r Recall) error {
	_, err := DB.Exec(`
		INSERT INTO recalls (
			id, rappel_guid, numero_fiche, numero_version, nature_juridique_rappel,
			categorie_produit, sous_categorie_produit, marque_produit, modeles_ou_references,
			identification_produits, conditionnements, date_debut_commercialisation,
			date_date_fin_commercialisation, date_limite_de_consommation, temperature_conservation,
			marque_salubrite, informations_complementaires, zone_geographique_de_vente,
			distributeurs, motif_rappel, risques_encourus, preconisations_sanitaires,
			description_complementaire_risque, conduites_a_tenir_par_le_consommateur,
			numero_contact, modalites_de_compensation, date_de_fin_de_la_procedure_de_rappel,
			informations_complementaires_publiques, liens_vers_les_images, lien_vers_la_liste_des_produits,
			lien_vers_la_liste_des_distributeurs, lien_vers_affichette_pdf, lien_vers_la_fiche_rappel,
			date_publication, libelle, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22,
			$23, $24,
			$25, $26, $27,
			$28, $29, $30,
			$31, $32, $33,
			$34, $35, now()
		)
		ON CONFLICT (id)
		DO UPDATE SET
			rappel_guid = EXCLUDED.rappel_guid,
			numero_fiche = EXCLUDED.numero_fiche,
			numero_version = EXCLUDED.numero_version,
			nature_juridique_rappel = EXCLUDED.nature_juridique_rappel,
			categorie_produit = EXCLUDED.categorie_produit,
			sous_categorie_produit = EXCLUDED.sous_categorie_produit,
			marque_produit = EXCLUDED.marque_produit,
			modeles_ou_references = EXCLUDED.modeles_ou_references,
			identification_produits = EXCLUDED.identification_produits,
			conditionnements = EXCLUDED.conditionnements,
			date_debut_commercialisation = EXCLUDED.date_debut_commercialisation,
			date_date_fin_commercialisation = EXCLUDED.date_date_fin_commercialisation,
			date_limite_de_consommation = EXCLUDED.date_limite_de_consommation,
			temperature_conservation = EXCLUDED.temperature_conservation,
			marque_salubrite = EXCLUDED.marque_salubrite,
			informations_complementaires = EXCLUDED.informations_complementaires,
			zone_geographique_de_vente = EXCLUDED.zone_geographique_de_vente,
			distributeurs = EXCLUDED.distributeurs,
			motif_rappel = EXCLUDED.motif_rappel,
			risques_encourus = EXCLUDED.risques_encourus,
			preconisations_sanitaires = EXCLUDED.preconisations_sanitaires,
			description_complementaire_risque = EXCLUDED.description_complementaire_risque,
			conduites_a_tenir_par_le_consommateur = EXCLUDED.conduites_a_tenir_par_le_consommateur,
			numero_contact = EXCLUDED.numero_contact,
			modalites_de_compensation = EXCLUDED.modalites_de_compensation,
			date_de_fin_de_la_procedure_de_rappel = EXCLUDED.date_de_fin_de_la_procedure_de_rappel,
			informations_complementaires_publiques = EXCLUDED.informations_complementaires_publiques,
			liens_vers_les_images = EXCLUDED.liens_vers_les_images,
			lien_vers_la_liste_des_produits = EXCLUDED.lien_vers_la_liste_des_produits,
			lien_vers_la_liste_des_distributeurs = EXCLUDED.lien_vers_la_liste_des_distributeurs,
			lien_vers_affichette_pdf = EXCLUDED.lien_vers_affichette_pdf,
			lien_vers_la_fiche_rappel = EXCLUDED.lien_vers_la_fiche_rappel,
			date_publication = EXCLUDED.date_publication,
			libelle = EXCLUDED.libelle,
			updated_at = now()
	`,
		r.ID, r.RappelGUID, r.NumeroFiche, r.NumeroVersion, r.NatureJuridiqueRappel,
		r.CategorieProduit, r.SousCategorieProduit, r.MarqueProduit, r.ModelesOuReferences,
		r.IdentificationProduits, r.Conditionnements, r.DateDebutCommercialisation,
		r.DateFinCommercialisation, r.DateLimiteDeConsommation, r.TemperatureConservation,
		r.MarqueSalubrite, r.InformationsComplementaires, r.ZoneGeographiqueDeVente,
		r.Distributeurs, r.MotifRappel, r.RisquesEncourus, r.PreconisationsSanitaires,
		r.DescriptionComplementaireRisque, r.ConduitesATenirParLeConsommateur,
		r.NumeroContact, r.ModalitesDeCompensation, r.DateDeFinDeLaProcedureDeRappel,
		r.InformationsComplementairesPubliques, r.LiensVersLesImagesRaw, r.LienVersLaListeDesProduits,
		r.LienVersLaListeDesDistributeurs, r.LienVersAffichettePDF, r.LienVersLaFicheRappel,
		r.DatePublication, r.Libelle,
	)
	return err
}

func GetPaginatedRecalls(page int, pageSize int) ([]Recall, error) {
	return GetPaginatedRecallsWithLimit(page, pageSize, pageSize)
}

func GetPaginatedRecallsWithLimit(page int, pageSize int, limit int) ([]Recall, error) {
	offset := (page - 1) * pageSize
	rows, err := DB.Query(`
		SELECT
			id,
			numero_fiche,
			categorie_produit,
			sous_categorie_produit,
			marque_produit,
			risques_encourus,
			motif_rappel,
			preconisations_sanitaires,
			numero_contact,
			distributeurs,
			modalites_de_compensation,
			zone_geographique_de_vente, 
			lien_vers_affichette_pdf,
			lien_vers_la_fiche_rappel,
			liens_vers_les_images,
			libelle,
			date_publication
		FROM recalls
		ORDER BY date_publication DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recalls []Recall
	for rows.Next() {
		var r Recall
		err := rows.Scan(
			&r.ID,
			&r.NumeroFiche,
			&r.CategorieProduit,
			&r.SousCategorieProduit,
			&r.MarqueProduit,
			&r.RisquesEncourus,
			&r.MotifRappel,
			&r.PreconisationsSanitaires,
			&r.NumeroContact,
			&r.Distributeurs,
			&r.ModalitesDeCompensation,
			&r.ZoneGeographiqueDeVente,
			&r.LienVersAffichettePDF,
			&r.LienVersLaFicheRappel,
			&r.LiensVersLesImagesRaw,
			&r.Libelle,
			&r.DatePublication,
		)
		if err != nil {
			return nil, err
		}
		r.ImageURLs = parseImageURLs(r.LiensVersLesImagesRaw)

		recalls = append(recalls, r)
	}

	return recalls, nil
}
func GetAllRisks() ([]string, error) {
	rows, err := DB.Query(`
        SELECT DISTINCT TRIM(risk) AS risk
        FROM recalls, unnest(string_to_array(risques_encourus, '|')) AS risk
        WHERE risques_encourus IS NOT NULL AND risques_encourus <> ''
        ORDER BY risk;
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var risks []string
	for rows.Next() {
		var risk string
		if err := rows.Scan(&risk); err != nil {
			return nil, err
		}
		risks = append(risks, risk)
	}
	return risks, nil
}

func GetAllCategories() ([]string, error) {
	rows, err := DB.Query(`SELECT DISTINCT categorie_produit FROM recalls ORDER BY categorie_produit`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		if cat != "" {
			categories = append(categories, cat)
		}
	}
	return categories, nil
}
func GetPaginatedRecallsByCategory(page int, pageSize int, category string) ([]Recall, error) {
	offset := (page - 1) * pageSize
	rows, err := DB.Query(`
		SELECT id, numero_fiche, categorie_produit, sous_categorie_produit,
		       marque_produit, risques_encourus, motif_rappel,
		       preconisations_sanitaires, numero_contact, distributeurs,
		       modalites_de_compensation, lien_vers_affichette_pdf,
		       lien_vers_la_fiche_rappel, liens_vers_les_images, libelle,
		       date_publication
		FROM recalls
		WHERE categorie_produit = $1
		ORDER BY date_publication DESC
		LIMIT $2 OFFSET $3
	`, category, pageSize, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recalls []Recall
	for rows.Next() {
		var r Recall
		err := rows.Scan(
			&r.ID, &r.NumeroFiche, &r.CategorieProduit, &r.SousCategorieProduit,
			&r.MarqueProduit, &r.RisquesEncourus, &r.MotifRappel,
			&r.PreconisationsSanitaires, &r.NumeroContact, &r.Distributeurs,
			&r.ModalitesDeCompensation, &r.LienVersAffichettePDF,
			&r.LienVersLaFicheRappel, &r.LiensVersLesImagesRaw, &r.Libelle,
			&r.DatePublication,
		)
		if err != nil {
			return nil, err
		}
		r.ImageURLs = parseImageURLs(r.LiensVersLesImagesRaw)

		recalls = append(recalls, r)
	}
	return recalls, nil
}
func GetAllZones() ([]string, error) {
	rows, err := DB.Query(`SELECT DISTINCT zone_geographique_de_vente FROM recalls WHERE zone_geographique_de_vente <> '' ORDER BY zone_geographique_de_vente`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []string
	for rows.Next() {
		var zone string
		if err := rows.Scan(&zone); err != nil {
			return nil, err
		}
		if zone != "" {
			zones = append(zones, zone)
		}
	}
	return zones, nil
}

func GetAllBrands() ([]string, error) {
	rows, err := DB.Query(`SELECT DISTINCT marque_produit FROM recalls WHERE marque_produit <> '' ORDER BY marque_produit`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var brand string
		if err := rows.Scan(&brand); err != nil {
			return nil, err
		}
		if brand != "" {
			brands = append(brands, brand)
		}
	}
	return brands, nil
}
func GetPaginatedRecallsFiltered(page, pageSize int, category, zone, risk, brand string, dateStart, dateEnd string) ([]Recall, error) {
	return GetPaginatedRecallsFilteredWithLimit(page, pageSize, pageSize, category, zone, risk, brand, dateStart, dateEnd)
}

func GetPaginatedRecallsFilteredWithLimit(page, pageSize, limit int, category, zone, risk, brand string, dateStart, dateEnd string) ([]Recall, error) {
	offset := (page - 1) * pageSize

	// Build dynamic query with parameters
	query := `
		SELECT
			id, numero_fiche, categorie_produit, sous_categorie_produit,
			marque_produit, risques_encourus, motif_rappel,
			preconisations_sanitaires, numero_contact, distributeurs,
			modalites_de_compensation, zone_geographique_de_vente,
			lien_vers_affichette_pdf, lien_vers_la_fiche_rappel,
			liens_vers_les_images, libelle, date_publication
		FROM recalls
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if category != "" {
		query += ` AND categorie_produit = $` + strconv.Itoa(argIdx)
		args = append(args, category)
		argIdx++
	}
	if zone != "" {
		query += ` AND zone_geographique_de_vente = $` + strconv.Itoa(argIdx)
		args = append(args, zone)
		argIdx++
	}
	if risk != "" {
		query += ` AND ('|' || risques_encourus || '|') ILIKE '%' || '|' || $` + strconv.Itoa(argIdx) + ` || '|' || '%'`
		args = append(args, risk)
		argIdx++
	}

	if brand != "" {
		query += ` AND marque_produit = $` + strconv.Itoa(argIdx)
		args = append(args, brand)
		argIdx++
	}
	if dateStart != "" {
		query += ` AND date_publication >= $` + strconv.Itoa(argIdx)
		args = append(args, dateStart)
		argIdx++
	}
	if dateEnd != "" {
		query += ` AND date_publication <= $` + strconv.Itoa(argIdx)
		args = append(args, dateEnd)
		argIdx++
	}

	query += ` ORDER BY date_publication DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recalls []Recall
	for rows.Next() {
		var r Recall
		err := rows.Scan(
			&r.ID, &r.NumeroFiche, &r.CategorieProduit, &r.SousCategorieProduit,
			&r.MarqueProduit, &r.RisquesEncourus, &r.MotifRappel,
			&r.PreconisationsSanitaires, &r.NumeroContact, &r.Distributeurs,
			&r.ModalitesDeCompensation, &r.ZoneGeographiqueDeVente,
			&r.LienVersAffichettePDF, &r.LienVersLaFicheRappel,
			&r.LiensVersLesImagesRaw, &r.Libelle, &r.DatePublication,
		)
		if err != nil {
			return nil, err
		}
		r.ImageURLs = parseImageURLs(r.LiensVersLesImagesRaw)
		recalls = append(recalls, r)
	}
	return recalls, nil
}
