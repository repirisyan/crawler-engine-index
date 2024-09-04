package mysqlTempItem

import (
	"crawler-index/db/mysql"
	"database/sql"
	"fmt"
)

// Category represents a category with its associated keywords and commodities.
type Category struct {
	Keyword                string  `json:"keyword"`
	SubComodity            *string `json:"sub_comodity"`
	SecondLevelSubComodity *string `json:"second_level_sub_comodity"`
	ThirdLevelSubComodity  *string `json:"third_level_sub_comodity"`
	Comodity               string  `json:"comodity"`
}

// GetAllData retrieves all categories and their related data from the database.
func GetCategory(keywordID uint64) (*Category, error) {
	query := `
		SELECT 
			k.name AS keyword, 
			k.sub_comodity, 
			k.second_level_sub_comodity, 
			k.third_level_sub_comodity, 
			c.name AS comodity 
		FROM keywords k
		INNER JOIN comodities c ON k.comodity_id = c.id
		WHERE k.id = ?`

	// Using QueryRow instead of Query since we expect only one result
	row := mysql.DB.QueryRow(query, keywordID)

	var category Category
	// Scan the result into the Category struct
	if err := row.Scan(&category.Keyword, &category.SubComodity, &category.SecondLevelSubComodity, &category.ThirdLevelSubComodity, &category.Comodity); err != nil {
		if err == sql.ErrNoRows {
			// Handle the case where no rows were returned
			return nil, fmt.Errorf("no category found with keyword ID %d", keywordID)
		}
		return nil, fmt.Errorf("failed to scan row for keyword ID %d: %w", keywordID, err)
	}

	return &category, nil
}
