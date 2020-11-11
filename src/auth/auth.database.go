package auth

import (
	"context"
	"database/sql"
	"time"

	"confusion.com/bwoo/database"
)

func createUserInDb(signupInfo UserInfo) (int64, bool) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(signupInfo.Password), costOfPwHash)
	hashedPasswd, err := signupInfo.generatePasswordHash()
	if err != nil {
		return userIdNotFound, false
	}

	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO user(
															firstname,
															lastname,
															username,
															password
														)
														VALUES (
															?,?,?,?
														)`,
		signupInfo.Firstname,
		signupInfo.Lastname,
		signupInfo.Username,
		hashedPasswd)
	if err != nil {
		return userIdNotFound, false
	}

	userId, _ := results.LastInsertId()
	return userId, true

}

func validateUserInDb(creds credentials) (int64, bool, bool) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT 
														id,
														password,
														admin
													FROM user
													WHERE username = ?`,
		creds.Username)

	var userId int64
	var passwdHash passwordHash
	var isAdmin bool
	err := row.Scan(&userId, &passwdHash, &isAdmin)
	if err == sql.ErrNoRows {
		return userIdNotFound, false, false
	} else if err != nil {
		return userIdNotFound, false, false
	}

	if err = passwdHash.validateCredentials(creds); err != nil {
		return userIdNotFound, false, false
	}

	return userId, isAdmin, true
}

func getUsersFromDb() ([]UserInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := database.DbConn.QueryContext(ctx, `SELECT id, 
														firstname, 
														lastname, 
														admin, 
														username,
														createdAt,
														UpdatedAt 
													FROM user`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	userInfoList := make([]UserInfo, 0)
	for rows.Next() {

		var user UserInfo
		rows.Scan(&user.ID,
			&user.Firstname,
			&user.Lastname,
			&user.Admin,
			&user.Username,
			&user.CreatedAt,
			&user.UpdatedAt)
		userInfoList = append(userInfoList, user)
	}

	return userInfoList, nil
}
