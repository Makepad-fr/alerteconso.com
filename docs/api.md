# AlerteConso API

Base URL: `https://alerteconso.com`

The public API is read-only and exposes resources directly under their semantic URIs. There is no `/api` namespace. JSON representations expose HATEOAS links in `_links` fields and RFC 8288 `Link` headers where pagination applies.

`/recalls` is the recall collection resource. Clients receive JSON by default and should send `Accept: application/json`. Browsers that send an HTML-preferred `Accept` header receive the existing web page for the same collection resource.

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

Supported by `GET /recalls`.

| Parameter | Default | Description |
| --- | --- | --- |
| `page` | `1` | 1-based page number |
| `pageSize` | `20` | Number of recalls per page |
| `q` | empty | Free-text search across label, brand, references, product identifiers, and recall number |
| `category` | empty | Exact `categorie_produit` filter |
| `zone` | empty | Exact `zone_geographique_de_vente` filter |
| `brand` | empty | Exact `marque_produit` filter |
| `risk` | empty | Pipe-delimited risk token filter |
| `dateStart` | empty | Inclusive `date_publication` lower bound, formatted as `YYYY-MM-DD` or RFC 3339 |
| `dateEnd` | empty | Inclusive `date_publication` upper bound, formatted as `YYYY-MM-DD` or RFC 3339 |

## Resources

### Service Entry Point

```http
GET /
Accept: application/json
```

Returns discoverable links for the public resources.

```json
{
  "_links": [
    { "rel": "self", "href": "/" },
    { "rel": "recalls", "href": "/recalls" },
    { "rel": "recall-filters", "href": "/recalls/filters" },
    { "rel": "recall-categories", "href": "/recalls/categories" },
    { "rel": "recall-risks", "href": "/recalls/risks" },
    { "rel": "recall-zones", "href": "/recalls/zones" },
    { "rel": "recall-brands", "href": "/recalls/brands" }
  ]
}
```

### Recall Collection

```http
GET /recalls?page=1&pageSize=20
Accept: application/json
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
        { "rel": "collection", "href": "/recalls" },
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
    { "rel": "self", "href": "/recalls?page=1&pageSize=20" },
    { "rel": "first", "href": "/recalls?page=1&pageSize=20" },
    { "rel": "next", "href": "/recalls?page=2&pageSize=20" }
  ]
}
```

`page.count` is the number of items returned in the current page, not the total number of matching recalls. The response also includes an HTTP `Link` header with the same collection navigation relations. The `next` relation is only emitted when another page currently exists.

### Recall Resource

```http
GET /recalls/{id}
Accept: application/json
```

Returns one recall resource with `_links`.

```json
{
  "id": 49788,
  "numero_fiche": "sr/01361/26",
  "libelle": "chaussures pour enfants",
  "_links": [
    { "rel": "self", "href": "/recalls/49788" },
    { "rel": "collection", "href": "/recalls" },
    { "rel": "official", "href": "https://rappel.conso.gouv.fr/fiche-rappel/49788/rapex" },
    { "rel": "pdf", "href": "https://rappel.conso.gouv.fr/affichettepdf/49788/rapex" }
  ]
}
```

### Recall Filter Metadata

```http
GET /recalls/filters
GET /recalls/categories
GET /recalls/risks
GET /recalls/zones
GET /recalls/brands
Accept: application/json
```

Filter responses contain `data`, `count`, and `_links`, except `/recalls/filters`, which groups all filter arrays in one response.

## HTML Representations

```http
GET /
GET /recalls
Accept: text/html
```

These requests return the browser interface for the service entry point and recall collection.

## Status Endpoints

```http
GET /healthz
GET /readyz
```

`/healthz` reports process liveness. `/readyz` checks the database connection.
