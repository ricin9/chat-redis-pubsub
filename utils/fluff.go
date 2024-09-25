package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"ricin9/fiber-chat/services"
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

func GetMembersByIds(rows *sql.Rows) (members []services.Member, err error) {
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		members = append(members, services.Member{ID: id, Username: services.GetUsername(context.Background(), id)})
	}
	return members, nil
}
