package users

import (
	"encoding/json"
	"net/http"

	"github.com/soumalya/food-delivery-admin/database"
	"github.com/soumalya/food-delivery-admin/model"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(r.Context(), `
		SELECT DISTINCT ON (u.user_id) 
            u.user_id, u.name, u.mobile_no, u.building_no, u.room_no, u.role, u.plan,
            COALESCE(w.balance_after, 0) AS balance
        FROM users u
        LEFT JOIN wallet_transactions w ON u.user_id = w.user_id
        ORDER BY u.user_id, w.created_at DESC;`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.UserID, &u.Name, &u.MobileNo, &u.BuildingNo, &u.RoomNo, &u.Role, &u.Plan, &u.Balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.Role == "" {
		u.Role = "normal"
	}

	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	err = tx.QueryRow(r.Context(), `
		INSERT INTO USERS (NAME, MOBILE_NO, BUILDING_NO, ROOM_NO, ROLE, PLAN) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING USER_ID
	`, u.Name, u.MobileNo, u.BuildingNo, u.RoomNo, u.Role, u.Plan).Scan(&u.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(r.Context(), `INSERT INTO WALLET (USER_ID, BALANCE) VALUES ($1, 0)`, u.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit(r.Context())
	json.NewEncoder(w).Encode(u)
}
