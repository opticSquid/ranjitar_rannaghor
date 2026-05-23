package users

import (
	"context"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func FetchUsersFromDB(ctx context.Context) ([]User, error) {
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, `
		SELECT DISTINCT ON (u.user_id)
            u.user_id, u.name, u.mobile_no, u.building_no, u.room_no, u.role, u.plan,
            COALESCE(w.balance_after, 0) AS balance
        FROM users u
        LEFT JOIN wallet_transactions w ON u.user_id = w.user_id
        ORDER BY u.user_id, w.created_at DESC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.UserID, &u.Name, &u.MobileNo, &u.BuildingNo, &u.RoomNo, &u.Role, &u.Plan, &u.Balance)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func InsertUserToDB(ctx context.Context, u *User) error {
	dbPool := database.GetDbConn()
	return dbPool.QueryRow(ctx, `
		INSERT INTO USERS (NAME, MOBILE_NO, BUILDING_NO, ROOM_NO, ROLE, PLAN)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING USER_ID
	`, u.Name, u.MobileNo, u.BuildingNo, u.RoomNo, u.Role, u.Plan).Scan(&u.UserID)
}
