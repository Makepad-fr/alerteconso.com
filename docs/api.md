# AlerteConso API

Base URL: `https://alerteconso.com`

The API is read-only for public clients and is scoped under `/api`. The `/recalls` route is the browser HTML page, not an API endpoint. API responses use JSON and expose HATEOAS links in `_links` fields and RFC 8288 `Link` headers.

## Link Relations

| Relation | Meaning |
| --- | --- |
| `self` | Current resource URL |
| `collection` | Parent collection URL |
| `first` | First page of a collection |
| `prev` | Previous page, when available |
| `next` | Next page, when available |
| `official` | Official RappelConso recall page |
| `pdf` | Official recall poster PDF |
| `recalls` | Recall collection |
| `recall-filters` | Combined filter metadata for recalls |
| `recall-categories` | Recall category values |
| `recall-risks` | Recall risk values |
| `recall-zones` | Recall sale-zone values |
| `recall-brands` | Recall brand values |

## Collection Query Parameters

Supported by `GET /api/recalls`.

| Parameter | Default | Description |
| --- | --- | --- |
| `page` | `1` | 1-based page number |
| `pageSize` | `20` | Number of recalls per page |
| `q` | empty | Free-text search across label, brand, references, product identifiers, and recall number |
| `category` | empty | Exact `categorie_produit` filter |
| `zone` | empty | Exact `zone_geographique_de_vente` filter |
| `brand` | empty | Exact `marque_produit` filter |
| `risk` | empty | Pipe-delimited risk token filter |
| `dateStart` | empty | Inclusive `date_publication` lower bound |
| `dateEnd` | empty | Inclusive `date_publication` upper bound |

## Endpoints

### API Root

```http
GET /api
```

Returns discoverable API links.

```json
{
  "_links": [
    { "rel": "self", "href": "/api" },
    { "rel": "recalls", "href": "/api/recalls" },
    { "rel": "recall-filters", "href": "/api/recalls/filters" },
    { "rel": "recall-categories", "href": "/api/recalls/categories" },
    { "rel": "recall-risks", "href": "/api/recalls/risks" },
    { "rel": "recall-zones", "href": "/api/recalls/zones" },
    { "rel": "recall-brands", "href": "/api/recalls/brands" }
  ]
}
```

### List Recalls

```http
GET /api/recalls?page=1&pageSize=20
```

Returns a REST collection envelope.

```json
{
  "data": [
    {
      "id": 49788,
      "numero_fiche": "sr/01361/26",
      "categorie_produit": "vêtements, mode, epi",
      "marque_produit": "xmgolong",
      "libelle": "chaussures pour enfants",
      "_links": [
        { "rel": "self", "href": "/api/recalls/49788" },
        { "rel": "collection", "href": "/api/recalls" },
        { "rel": "official", "href": "https://rappel.conso.gouv.fr/fiche-rappel/49788/rapex" }
      ]
    }
  ],
  "page": {
    "page": 1,
    "pageSize": 20,
    "count": 1
  },
  "_links": [
    { "rel": "self", "href": "/api/recalls?page=1&pageSize=20" },
    { "rel": "first", "href": "/api/recalls?page=1&pageSize=20" },
    { "rel": "next", "href": "/api/recalls?page=2&pageSize=20" }
  ]
}
```

The response also includes an HTTP `Link` header with the same collection navigation relations.

### Get Recall

```http
GET /api/recalls/{id}
```

Returns one recall resource with `_links`.

```json
{
  "id": 49788,
  "numero_fiche": "sr/01361/26",
  "libelle": "chaussures pour enfants",
  "_links": [
    { "rel": "self", "href": "/api/recalls/49788" },
    { "rel": "collection", "href": "/api/recalls" },
    { "rel": "official", "href": "https://rappel.conso.gouv.fr/fiche-rappel/49788/rapex" },
    { "rel": "pdf", "href": "https://rappel.conso.gouv.fr/affichettepdf/49788/rapex" }
  ]
}
```

### Recall Filter Metadata

```http
GET /api/recalls/filters
GET /api/recalls/categories
GET /api/recalls/risks
GET /api/recalls/zones
GET /api/recalls/brands
```

Filter responses contain `data`, `count`, and `_links`, except `/api/recalls/filters`, which groups all filter arrays in one response.

## Web Routes

```http
GET /
GET /recalls
```

These routes return HTML and are not part of the JSON API.

## Status Endpoints

```http
GET /healthz
GET /readyz
```

`/healthz` reports process liveness. `/readyz` checks the database connection.
