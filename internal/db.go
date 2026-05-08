package internal

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error

	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				err = ensureRecallDatePublicationTimestamptz(DB)
				if err == nil {
					fmt.Println("✅ Connected to PostgreSQL")
					return
				}
			}
		}

		log.Printf("⏳ Waiting for DB... (%d/10) error: %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("❌ Failed to connect to DB after retries: %v", err)
}

func ensureRecallDatePublicationTimestamptz(db *sql.DB) error {
	var dataType string
	err := db.QueryRow(`
		SELECT data_type
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'recalls'
			AND column_name = 'date_publication'
	`).Scan(&dataType)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if dataType == "timestamp with time zone" {
		return nil
	}
	if dataType != "timestamp without time zone" {
		return fmt.Errorf("unsupported recalls.date_publication type %q", dataType)
	}

	_, err = db.Exec(`
		ALTER TABLE recalls
		ALTER COLUMN date_publication TYPE TIMESTAMPTZ
		USING date_publication AT TIME ZONE 'UTC'
	`)
	return err
}

func GetRecallByID(id int) (Recall, error) {
	var r Recall
	err := DB.QueryRow(`
		SELECT
			id,
			rappel_guid,
			numero_fiche,
			numero_version,
			nature_juridique_rappel,
			categorie_produit,
			sous_categorie_produit,
			marque_produit,
			modeles_ou_references,
			identification_produits,
			conditionnements,
			date_debut_commercialisation,
			date_date_fin_commercialisation,
			date_limite_de_consommation,
			temperature_conservation,
			marque_salubrite,
			informations_complementaires,
			zone_geographique_de_vente,
			distributeurs,
			motif_rappel,
			risques_encourus,
			preconisations_sanitaires,
			description_complementaire_risque,
			conduites_a_tenir_par_le_consommateur,
			numero_contact,
			modalites_de_compensation,
			date_de_fin_de_la_procedure_de_rappel,
			informations_complementaires_publiques,
			liens_vers_les_images,
			lien_vers_la_liste_des_produits,
			lien_vers_la_liste_des_distributeurs,
			lien_vers_affichette_pdf,
			lien_vers_la_fiche_rappel,
			`+datePublicationRFC3339SQL+` AS date_publication,
			libelle
		FROM recalls
		WHERE id = $1
	`, id).Scan(
		&r.ID,
		&r.RappelGUID,
		&r.NumeroFiche,
		&r.NumeroVersion,
		&r.NatureJuridiqueRappel,
		&r.CategorieProduit,
		&r.SousCategorieProduit,
		&r.MarqueProduit,
		&r.ModelesOuReferences,
		&r.IdentificationProduits,
		&r.Conditionnements,
		&r.DateDebutCommercialisation,
		&r.DateFinCommercialisation,
		&r.DateLimiteDeConsommation,
		&r.TemperatureConservation,
		&r.MarqueSalubrite,
		&r.InformationsComplementaires,
		&r.ZoneGeographiqueDeVente,
		&r.Distributeurs,
		&r.MotifRappel,
		&r.RisquesEncourus,
		&r.PreconisationsSanitaires,
		&r.DescriptionComplementaireRisque,
		&r.ConduitesATenirParLeConsommateur,
		&r.NumeroContact,
		&r.ModalitesDeCompensation,
		&r.DateDeFinDeLaProcedureDeRappel,
		&r.InformationsComplementairesPubliques,
		&r.LiensVersLesImagesRaw,
		&r.LienVersLaListeDesProduits,
		&r.LienVersLaListeDesDistributeurs,
		&r.LienVersAffichettePDF,
		&r.LienVersLaFicheRappel,
		&r.DatePublication,
		&r.Libelle,
	)
	return r, err
}
