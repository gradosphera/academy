package types

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Time wraps sql.NullTime with custom JSON marshalling.
type Time struct {
	sql.NullTime
}

func NewTime(t time.Time) Time {
	return Time{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: true,
		},
	}
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var dest *time.Time

	if err := json.Unmarshal(data, &dest); err != nil {
		return err
	}

	if dest == nil {
		t.Valid = false

		return nil
	}

	t.Time = *dest
	t.Valid = true

	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(t.Time)
}

func (t Time) IsEqual(t2 Time) bool {
	if !t.Valid && !t2.Valid {
		return true
	}

	if !t.Valid || !t2.Valid {
		return false
	}

	return t.Time.Equal(t2.Time)
}
