package models

import (
	"context"

	pgdb "crawler-index/db/postgres"
)

var Ctx = context.Background()

// GetKeywordComodityMap returns a keyword_id -> comodity_id lookup for every
// keyword. Used by category validation to keep comodity_id consistent with a
// re-predicted keyword_id.
func GetKeywordComodityMap() (map[uint64]uint64, error) {
	conn := pgdb.GetPostgresPool()

	rows, err := conn.Query(Ctx, `SELECT id, comodity_id FROM keywords`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[uint64]uint64)
	for rows.Next() {
		var id, comodityID uint64
		if err := rows.Scan(&id, &comodityID); err != nil {
			return nil, err
		}
		out[id] = comodityID
	}
	return out, rows.Err()
}
