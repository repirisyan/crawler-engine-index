package models

import (
	"context"
	pgdb "crawler-index/db/postgres"
)

var Ctx = context.Background() // Ensure you have a context here

type Brand struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	Product_origin bool   `json:"product_origin"`
}

func GetAllBrand(offset int, limit int) ([]Brand, error) {
	conn := pgdb.GetPostgresPool()

	query := `SELECT id, name, COALESCE(product_origin, false) FROM brands LIMIT $2 OFFSET $1`

	rows, err := conn.Query(Ctx, query, offset, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var brands []Brand

	for rows.Next() {
		var brand Brand
		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Product_origin,
		)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}
