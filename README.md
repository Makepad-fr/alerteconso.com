# AlerteConso - Product Recalls
<p align="center"><img src="./logo.png" alt="AlerteConso" width="120" /></p>

![Go](https://img.shields.io/badge/Go-1.24-blue) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-%23336791) ![Docker Compose](https://img.shields.io/badge/Docker-Compose-2496ED) ![License](https://img.shields.io/badge/License-MIT-green)


**AlerteConso** is a fast, responsive interface to explore French **product recalls** (official **RappelConso/DGCCRF** data).  
Filters by **category, risk, zone, brand**, **date range**, and links to **official pages/PDFs**.

- Endpoints: `/` (HTML), `/recalls` (HTML), `/api/recalls` (JSON), `/healthz`, `/readyz`
- Quickstart: `docker compose up -d --build` → open `http://localhost:9092`

## API

See [docs/api.md](docs/api.md) for the REST API contract, HATEOAS links, query parameters, and examples.

*Not affiliated with DGCCRF.*
