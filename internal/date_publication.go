package internal

// The database stores date_publication as a timestamp. Project it with an
// explicit format so API responses do not depend on PostgreSQL text rendering.
const datePublicationRFC3339SQL = `COALESCE(to_char(date_publication, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`
