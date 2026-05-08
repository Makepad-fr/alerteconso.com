package internal

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FlexibleText string

func (t *FlexibleText) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*t = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*t = FlexibleText(s)
		return nil
	}

	var values []any
	if err := json.Unmarshal(data, &values); err == nil {
		parts := make([]string, 0, len(values))
		for _, value := range values {
			part := jsonValueToString(value)
			if part != "" {
				parts = append(parts, part)
			}
		}
		*t = FlexibleText(strings.Join(parts, " | "))
		return nil
	}

	return fmt.Errorf("unsupported text JSON value: %s", raw)
}

func (t *FlexibleText) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*t = ""
	case string:
		*t = FlexibleText(v)
	case []byte:
		*t = FlexibleText(string(v))
	case time.Time:
		*t = FlexibleText(v.Format(time.RFC3339))
	default:
		*t = FlexibleText(fmt.Sprint(v))
	}
	return nil
}

func (t FlexibleText) Value() (driver.Value, error) {
	return string(t), nil
}

func (t FlexibleText) String() string {
	return string(t)
}

type FlexibleLinks []Link

func (links *FlexibleLinks) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*links = nil
		return nil
	}

	var list []Link
	if err := json.Unmarshal(data, &list); err == nil {
		*links = FlexibleLinks(list)
		return nil
	}

	var single Link
	if err := json.Unmarshal(data, &single); err == nil && (single.Rel != "" || single.Href != "") {
		*links = FlexibleLinks{single}
		return nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err == nil && object != nil && len(object) == 0 {
		*links = nil
		return nil
	}

	return fmt.Errorf("unsupported links JSON value: %s", raw)
}

func jsonValueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}
