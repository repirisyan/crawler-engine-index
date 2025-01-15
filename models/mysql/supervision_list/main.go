package mysqlSupervisionList

import (
	"crawler-index/db/mysql"
)

type Supervision struct {
	Name     string
	Category string
}

func GetAllData() ([]Supervision, error) {
	query := "SELECT name from supervision_lists WHERE status = 1"
	rows, err := mysql.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var superVisionList []Supervision

	for rows.Next() {
		var data Supervision
		err := rows.Scan(&data.Name)
		if err != nil {
			return nil, err
		}
		superVisionList = append(superVisionList, data)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return superVisionList, nil
}
