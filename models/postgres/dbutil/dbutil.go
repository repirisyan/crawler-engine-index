// Package dbutil holds small helpers shared by the postgres model packages.
package dbutil

import "unicode/utf8"

// Truncate returns s clipped to at most max runes. Use it to keep scraped
// strings within the column's varchar(n) limit so one oversized value does not
// abort a whole insert batch.
func Truncate(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

// TruncatePtr is Truncate for a nullable string column. nil stays nil.
func TruncatePtr(s *string, max int) *string {
	if s == nil {
		return nil
	}
	t := Truncate(*s, max)
	return &t
}
