package internal

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var recallDateLayouts = []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"}

// date_publication is stored as TIMESTAMPTZ. Project it through UTC explicitly
// so API responses do not depend on the database session time zone.
const datePublicationRFC3339SQL = `COALESCE(to_char(date_publication AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')`

func normalizeDatePublicationForDB(value string) (sql.NullTime, error) {
	parsed, ok, err := parseRecallDate("date_publication", value)
	if err != nil {
		return sql.NullTime{}, err
	}
	if !ok {
		return sql.NullTime{}, nil
	}
	return sql.NullTime{Time: parsed.UTC(), Valid: true}, nil
}

func normalizeDateFilters(dateStart, dateEnd string) (sql.NullTime, sql.NullTime, error) {
	start, hasStart, err := parseRecallDate("dateStart", dateStart)
	if err != nil {
		return sql.NullTime{}, sql.NullTime{}, err
	}
	end, hasEnd, err := parseRecallDate("dateEnd", dateEnd)
	if err != nil {
		return sql.NullTime{}, sql.NullTime{}, err
	}
	if hasStart && hasEnd && start.After(end) {
		return sql.NullTime{}, sql.NullTime{}, requestValidationError("dateStart must be before or equal to dateEnd")
	}

	var startFilter sql.NullTime
	if hasStart {
		startFilter = sql.NullTime{Time: start.UTC(), Valid: true}
	}
	var endFilter sql.NullTime
	if hasEnd {
		endFilter = sql.NullTime{Time: end.UTC(), Valid: true}
	}
	return startFilter, endFilter, nil
}

func validateDateFilters(dateStart, dateEnd string) error {
	_, _, err := normalizeDateFilters(dateStart, dateEnd)
	return err
}

func parseRecallDate(name, value string) (time.Time, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, nil
	}
	for _, layout := range recallDateLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true, nil
		}
	}
	return time.Time{}, false, requestValidationError(fmt.Sprintf("%s must use YYYY-MM-DD or RFC3339 format", name))
}
