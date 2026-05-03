package db

import (
	"database/sql"
	"fmt"
	"time"
)

const timeLayout = time.RFC3339Nano

func encodeTime(t time.Time) string {
	return t.UTC().Format(timeLayout)
}

func decodeTime(s string) (time.Time, error) {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		// Fallback: SQLite default datetime format
		t, err = time.Parse("2006-01-02 15:04:05", s)
		if err != nil {
			return time.Time{}, fmt.Errorf("db: parse time %q: %w", s, err)
		}
	}
	return t.UTC(), nil
}

func decodeNullTime(ns sql.NullString) (*time.Time, error) {
	if !ns.Valid {
		return nil, nil
	}
	t, err := decodeTime(ns.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// func encodeNullTime(t *time.Time) sql.NullString {
// 	if t == nil {
// 		return sql.NullString{}
// 	}
// 	return sql.NullString{String: encodeTime(*t), Valid: true}
// }
