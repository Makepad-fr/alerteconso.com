# AlerteConso API

Base URL: `https://alerteconso.com`

The API is read-only for public clients. Responses use JSON and expose HATEOAS links in `_links` fields or RFC 8288 `Link` headers.

## Compatibility

`GET /recalls` remains the legacy mobile API endpoint. It returns JSON unless the request explicitly asks for an HTML page with `Accept: text/html`.

Examples:

```http
GET /recalls
Accept: */*
```

returns the legacy JSON array.

```http
GET /recalls
Accept: text/html
```

returns the browser HTML page.

New integrations should prefer the canonical `/api` routes because their collection responses include pagination metadata and body-level HATEOAS links.

## Link Relations

| Relation | Meaning |
| --- | --- |
| `self` | Current resource URL |
| `api` | Canonical API URL for a recall resource |
| `collection` | Legacy recalls collection |
| `api-collection` | Canonical recalls collection |
| `first` | First page of a collection |
| `prev` | Previous page, when available |
| `next` | Next page, when available |
| `official` | Official RappelConso recall page |
| `pdf` | Official recall poster PDF |

## Collection Query Parameters

Supported by `GET /recalls` and `GET /api/recalls`.

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
    { "rel": "filters", "href": "/api/filters" }
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
        { "rel": "self", "href": "/recalls/49788" },
        { "rel": "api", "href": "/api/recalls/49788" },
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

### Legacy List Recalls

```http
GET /recalls?page=1&pageSize=20
Accept: application/json
```

Returns the legacy JSON array for existing mobile clients. Each recall object still includes `_links`.

### Get Recall

```http
GET /api/recalls/{id}
```

Canonical detail endpoint.

```http
GET /recalls/{id}
GET /recall/{id}
```

Legacy aliases. All three return the same recall representation with `_links`.

### Filters

```http
GET /api/filters
GET /api/categories
GET /api/risks
GET /api/zones
GET /api/brands
```

Legacy aliases are also available without the `/api` prefix:

```http
GET /filters
GET /categories
GET /risks
GET /zones
GET /brands
```

Filter responses contain `data`, `count`, and `_links`.

## Status Endpoints

```http
GET /healthz
GET /readyz
```

`/healthz` reports process liveness. `/readyz` checks the database connection.
