package types

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
)

// Interval wraps pgtype.Interval with custom JSON marshalling.
type Interval struct {
	pgtype.Interval
}

func NewInterval(jsonInterval JsonInterval) Interval {
	var interval Interval
	interval.Microseconds = jsonInterval.Microseconds +
		1_000_000*jsonInterval.Seconds +
		60_000_000*jsonInterval.Minutes +
		3_600_000_000*jsonInterval.Hours
	interval.Days = jsonInterval.Days
	interval.Months = jsonInterval.Months + 12*jsonInterval.Years
	interval.Valid = true

	return interval
}

type JsonInterval struct {
	Years        int32 `json:"years,omitempty"`
	Months       int32 `json:"months,omitempty"`
	Days         int32 `json:"days,omitempty"`
	Hours        int64 `json:"hours,omitempty"`
	Minutes      int64 `json:"minutes,omitempty"`
	Seconds      int64 `json:"seconds,omitempty"`
	Microseconds int64 `json:"microseconds,omitempty"`
}

func (in *Interval) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "null", `""`:
		return nil
	}

	var jsonInterval JsonInterval
	if err := json.Unmarshal(data, &jsonInterval); err != nil {
		return err
	}
	in.Interval.Microseconds = jsonInterval.Microseconds +
		1_000_000*jsonInterval.Seconds +
		60_000_000*jsonInterval.Minutes +
		3_600_000_000*jsonInterval.Hours
	in.Interval.Days = jsonInterval.Days
	in.Interval.Months = jsonInterval.Months + 12*jsonInterval.Years
	in.Interval.Valid = true

	return nil
}

func (in *Interval) MarshalJSON() ([]byte, error) {
	if !in.Valid {
		return json.Marshal(nil)
	}

	var ret JsonInterval

	ret.Years = in.Months / 12
	ret.Months = in.Months % 12
	ret.Days = in.Days
	ret.Hours = in.Microseconds / (60 * 60 * 1_000_000)

	remaining := in.Microseconds % (60 * 60 * 1_000_000)
	ret.Minutes = remaining / (60 * 1_000_000)

	remaining = remaining % (60 * 1_000_000)
	ret.Seconds = remaining / 1_000_000

	ret.Microseconds = remaining % 1_000_000

	return json.Marshal(ret)
}

func (in Interval) IsEqual(in2 Interval) bool {
	if !in.Valid && !in2.Valid {
		return true
	}

	if in.Valid != in2.Valid {
		return false
	}

	return in.Microseconds == in2.Microseconds &&
		in.Days == in2.Days &&
		in.Months == in2.Months
}
