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
const datePublicationRFC3339SQL = `to_char(date_publication AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')`

type dateFilterBound struct {
	Time      time.Time
	Valid     bool
	Exclusive bool
}

func normalizeDatePublicationForDB(value string) (sql.NullTime, error) {
	parsed, ok, _, err := parseRecallDate("date_publication", value)
	if err != nil {
		return sql.NullTime{}, err
	}
	if !ok {
		return sql.NullTime{}, requestValidationError("date_publication is required")
	}
	return sql.NullTime{Time: parsed.UTC(), Valid: true}, nil
}

func normalizeDateFilters(dateStart, dateEnd string) (dateFilterBound, dateFilterBound, error) {
	start, hasStart, _, err := parseRecallDate("dateStart", dateStart)
	if err != nil {
		return dateFilterBound{}, dateFilterBound{}, err
	}
	end, hasEnd, endDateOnly, err := parseRecallDate("dateEnd", dateEnd)
	if err != nil {
		return dateFilterBound{}, dateFilterBound{}, err
	}
	if hasEnd && endDateOnly {
		end = end.AddDate(0, 0, 1)
	}
	if hasStart && hasEnd {
		invalidRange := start.After(end)
		if endDateOnly {
			invalidRange = !start.Before(end)
		}
		if invalidRange {
			return dateFilterBound{}, dateFilterBound{}, requestValidationError("dateStart must be before or equal to dateEnd")
		}
	}

	var startFilter dateFilterBound
	if hasStart {
		startFilter = dateFilterBound{Time: start.UTC(), Valid: true}
	}
	var endFilter dateFilterBound
	if hasEnd {
		endFilter = dateFilterBound{Time: end.UTC(), Valid: true, Exclusive: endDateOnly}
	}
	return startFilter, endFilter, nil
}

func validateDateFilters(dateStart, dateEnd string) error {
	_, _, err := normalizeDateFilters(dateStart, dateEnd)
	return err
}

func parseRecallDate(name, value string) (time.Time, bool, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, false, nil
	}
	for _, layout := range recallDateLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true, layout == "2006-01-02", nil
		}
	}
	return time.Time{}, false, false, requestValidationError(fmt.Sprintf("%s must use YYYY-MM-DD or RFC3339 format", name))
}
