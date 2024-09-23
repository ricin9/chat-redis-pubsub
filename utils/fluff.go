package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func Dump(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Print(string(b))
}

/* dumb sh*t below */

func GetUserIds(rows *sql.Rows) (ids []int, err error) {
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}
	return ids, nil
}
