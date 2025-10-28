docker-compose build
docker-compose up

# AlerteConso — Rappels de produits (non officiel)
<p align="center"><img src="./logo.png" alt="AlerteConso" width="120" /></p>

![Go](https://img.shields.io/badge/Go-1.24-blue) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-%23336791) ![Docker Compose](https://img.shields.io/badge/Docker-Compose-2496ED) ![License](https://img.shields.io/badge/License-MIT-green)

> **AlerteConso** est une interface rapide et responsive pour explorer les **rappels de produits en France** (source officielle **RappelConso / DGCCRF** via data.gouv.fr).  
> Recherche par **catégorie, risque, zone, marque**, filtrage par **période**, et **liens vers les fiches/PDF officiels**.

## Fonctionnalités

-  **Back‑end Go** ultra rapide + **PostgreSQL 17**
-  **UI responsive** (mobile/tablette/desktop)
  - Mobile : bouton *Filtres* simple (non‑sticky)
  - Desktop : barre compacte + **tiroir latéral** des filtres
-  Filtres : catégorie, risque, zone, marque, **dates**
-  Pagination côté serveur
-  **Actualisation régulière** des rappels (fetch + upsert)
-  Endpoints **HTML & JSON** prêts à consommer
-  Mode clair/sombre, focus sur **accessibilité**

## Stack

- **Go** (net/http, templates)
- **PostgreSQL 17**
- **Docker / Docker Compose**
- HTML/CSS vanilla (pas de framework front)

---

## Démarrage rapide

Prérequis : Docker & Docker Compose

```bash
git clone https://github.com/votre-org/alerteconso.com.git
cd alerteconso.com
docker compose up -d --build
# Ouvrir l'app :
open http://localhost:9092
```

Par défaut :
- Application : `http://localhost:9092`
- Base de données : `localhost:5433` (PostgreSQL)
- La base est initialisée via `schema.sql`
- Le fetch des rappels démarre automatiquement (tâche en arrière‑plan)

### Variables / Configuration

`docker-compose.yml` (extrait) :
```yaml
services:
  rappel-api:
    build: .
    ports:
      - "9092:8080"
    environment:
      - DATABASE_URL=postgres://rappeluser:rappelpass@db:5432/rappeldb?sslmode=disable

  db:
    image: postgres:17
    ports:
      - "5433:5432"
    environment:
      POSTGRES_USER: rappeluser
      POSTGRES_PASSWORD: rappelpass
      POSTGRES_DB: rappeldb
```

---

## 🔗 Endpoints utiles

- `GET /` — liste HTML avec filtres, pagination
- `GET /recalls` — **JSON** (mêmes filtres via query params)
- `GET /recall/{id}` — détail (HTML)
- `GET /categories` / `risks` / `zones` / `brands` — listes (JSON/HTML)
- `GET /healthz` — *ok* (rapide, sans DB)
- `GET /readyz` — *readiness* (peut vérifier la DB)
- `GET /__ping` — *pong* (debug)

**Exemples**  
```bash
# Santé
curl http://localhost:9092/healthz
curl http://localhost:9092/__ping

# JSON avec filtres
curl "http://localhost:9092/recalls?category=alimentation&risk=salmonella&page=1"
```

---

## 🛠️ Développement

### Lancer en local (sans Docker)

1) Démarrer Postgres 17 et créer la base, ou utiliser Docker pour la DB uniquement :
```bash
docker compose up -d db
```

2) Exporter la variable d’environnement :
```bash
export DATABASE_URL="postgres://rappeluser:rappelpass@localhost:5433/rappeldb?sslmode=disable"
```

3) Lancer l’app :
```bash
go run main.go
# http://localhost:8080
```

---

## 🧪 Dépannage (FAQ)

- **`curl http://localhost:9092/healthz` timeout**
  - Vérifier que l’app écoute sur `0.0.0.0:8080` (et non `127.0.0.1`).
  - `docker compose logs -f rappel-api` pour voir les logs.
  - Forcer IPv4 : `curl --ipv4 -v http://localhost:9092/healthz`
  - Tester dans le conteneur (Alpine a `wget`) :
    ```bash
    docker compose exec rappel-api sh -lc 'wget -qO- http://127.0.0.1:8080/healthz || echo "wget exit $?"'
    ```
- **La page `/` est lente**  
  Généralement dû à une requête ou un rendu lourd — tester `/__ping` et `/healthz` (doivent répondre instantanément).

---

## Mentions & Données

- **Sources** : RappelConso / DGCCRF via data.gouv.fr  
- **Avertissement** : ce projet est **indépendant et non affilié** à la DGCCRF.  
- Les données restent la propriété de leurs détenteurs. Respectez les CGU de data.gouv.fr.

---

## Roadmap 

- Export CSV/JSON filtré
- Filtres multi‑sélections + chips
- Tri (date, gravité, catégorie)
- Notifications (email/webhook)
- Tests e2e & CI
- i18n / multi‑langues
- Accessibilité renforcée (labels, focus states)

---

## Contribuer

Les PR sont bienvenues : corrections, docs, améliorations UI/UX.  
Merci d’ouvrir une issue avant une refonte majeure.

---

## Licence

[MIT](LICENSE)

---

## English (short)

**AlerteConso** is a fast, responsive interface to explore French **product recalls** (official **RappelConso/DGCCRF** data).  
Filters by **category, risk, zone, brand**, **date range**, and links to **official pages/PDFs**.

- Stack: Go, PostgreSQL 17, Docker Compose
- Endpoints: `/` (HTML), `/recalls` (JSON), `/healthz`, `/readyz`, `/__ping`
- Quickstart: `docker compose up -d --build` → open `http://localhost:9092`

*Not affiliated with DGCCRF.*