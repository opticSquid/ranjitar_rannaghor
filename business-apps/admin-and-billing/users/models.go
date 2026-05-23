package users

type User struct {
	UserID     int     `json:"user_id"`
	Name       string  `json:"name"`
	MobileNo   string  `json:"mobile_no"`
	BuildingNo string  `json:"building_no"`
	RoomNo     string  `json:"room_no"`
	Role       string  `json:"role"`
	Plan       string  `json:"plan"`
	Balance    float64 `json:"balance"`
}
