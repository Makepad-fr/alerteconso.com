package internal

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// date_publication is normalized to UTC before insert. Project it with an
// explicit RFC3339 format so responses do not depend on PostgreSQL text output.
const datePublicationRFC3339SQL = `COALESCE(to_char(date_publication, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`

func normalizeDatePublicationForDB(value string) (sql.NullTime, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullTime{}, nil
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return sql.NullTime{Time: parsed.UTC(), Valid: true}, nil
		}
	}

	return sql.NullTime{}, fmt.Errorf("date_publication must use YYYY-MM-DD or RFC3339 format")
}
