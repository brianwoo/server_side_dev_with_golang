package oauth2

import (
	"context"
	"database/sql"
	"time"

	"confusion.com/bwoo/auth"
	"confusion.com/bwoo/database"
)

func createUserInDb(userInfo FacebookUserInfo) (int64, bool) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO user(
															facebookId,
															firstname,
															lastname,
															username
														)
														VALUES (
															?,?,?,?
														)`,
		userInfo.ID,
		userInfo.FirstName,
		userInfo.LastName,
		userInfo.Name)
	if err != nil {
		return userIdNotFound, false
	}

	userId, _ := results.LastInsertId()
	return userId, true
}

func getFacebookUserFromDb(facebookId string) (*auth.UserInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT 
													id,
													admin,
													firstname,
													lastname,													
													createdAt,
													updatedAt
												FROM user
												WHERE facebookId = ?`, facebookId)

	var userInfo auth.UserInfo
	err := row.Scan(&userInfo.ID,
		&userInfo.Admin,
		&userInfo.Firstname,
		&userInfo.Lastname,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
