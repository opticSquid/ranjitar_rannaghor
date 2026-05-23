package users

import (
	"context"
)

func GetAllUsersService(ctx context.Context) ([]User, error) {
	return FetchUsersFromDB(ctx)
}

func CreateUserService(ctx context.Context, u *User) error {
	if u.Role == "" {
		u.Role = "normal"
	}
	return InsertUserToDB(ctx, u)
}
