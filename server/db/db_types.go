package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type DateTime time.Time

// Time converts a DateTime to a time.Time.
func (d DateTime) Time() time.Time {
	return time.Time(d)
}

func (d *DateTime) Scan(src any) error {
	switch src := src.(type) {
	case time.Time:
		*d = DateTime(src)
		return nil
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", src)
		if err != nil {
			return fmt.Errorf("parsing time: %w", err)
		}
		*d = DateTime(t)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

func (d DateTime) Value() (driver.Value, error) {
	return time.Time(d), nil
}
