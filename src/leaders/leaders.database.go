package leaders

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"confusion.com/bwoo/database"
	"confusion.com/bwoo/misc"
)

func createLeaderInDb(leader Leader) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	featured, _ := strconv.ParseBool(*leader.Featured)
	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO leader(
															name,
															image,															
															designation,
															abbr,
															featured,
															description
														)
														VALUES (
															?,?,?,?,?,?
														)`,
		leader.Name,
		leader.Image,
		leader.Designation,
		leader.Abbr,
		featured,
		leader.Description)

	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsInserted, _ := results.RowsAffected()
	status.SetStatus(numRowsInserted, 1)
	return status, nil
}

func deleteLeadersFromDb() (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status := &misc.Status{}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM leader`)
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsDeleted, err := results.RowsAffected()
	status.SetStatus(numRowsDeleted, 1)
	return status, nil
}

func deleteLeaderFromDb(leaderId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM leader
														WHERE ID = ?`,
		leaderId)
	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsDeleted, err := results.RowsAffected()
	status.SetStatus(numRowsDeleted, 1)
	return status, nil
}

func buildUpdateSQLFromInput(leaderId int64, leader Leader) (string, []interface{}) {

	var sb strings.Builder
	args := make([]interface{}, 0)
	sb.WriteString("UPDATE leader SET ")

	if leader.Name != nil {
		sb.WriteString("name = ? ")
		args = append(args, *leader.Name)
	}

	if leader.Image != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("image = ? ")
		args = append(args, *leader.Image)
	}

	if leader.Designation != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("designation = ? ")
		args = append(args, *leader.Designation)
	}

	if leader.Abbr != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("abbr = ? ")
		args = append(args, *leader.Abbr)
	}

	if leader.Featured != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		featured, _ := strconv.ParseBool(*leader.Featured)
		sb.WriteString("featured = ? ")
		args = append(args, featured)
	}

	if leader.Description != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("description = ? ")
		args = append(args, *leader.Description)
	}

	// It seems we have nothing to update, just return
	// empty SQL and nil args
	if len(args) == 0 {
		return "", nil
	}

	sb.WriteString("WHERE id = ?")
	args = append(args, leaderId)
	return sb.String(), args
}

func updateLeaderFromDb(leaderId int64, leader Leader) (*Leader, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateSql, updateArgs := buildUpdateSQLFromInput(leaderId, leader)
	results, err := database.DbConn.ExecContext(ctx, updateSql, updateArgs...)
	if err != nil {
		log.Println("Error updating record ", leaderId)
		return nil, err
	}

	numRowsUpdated, _ := results.RowsAffected()
	if numRowsUpdated == 0 {
		return &Leader{}, fmt.Errorf("No rows updated")
	}

	leaderUpdated, err := getLeaderFromDb(leaderId)
	return leaderUpdated, err
}

func getLeaderFromDb(leaderId int64) (*Leader, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT
													id,
													name,
													image,
													designation,
													abbr,
													CASE WHEN featured = 0 then 'false' ELSE 'true' END,
													description,
													createdAt,
													updatedAt
												FROM leader
												WHERE id = ?`, leaderId)

	var leader Leader
	err := row.Scan(&leader.ID,
		&leader.Name,
		&leader.Image,
		&leader.Designation,
		&leader.Abbr,
		&leader.Featured,
		&leader.Description,
		&leader.CreatedAt,
		&leader.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &leader, nil
}

func getLeadersFromDb(isFeatured bool) ([]Leader, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sqlGetLeaders := `SELECT 
							id,
							name,
							image,
							designation,
							abbr,
							CASE WHEN featured = 0 then 'false' ELSE 'true' END,
							description,
							createdAt,
							updatedAt
						FROM leader`

	if isFeatured {
		sqlGetLeaders += " WHERE featured = 1"
	}

	rows, err := database.DbConn.QueryContext(ctx, sqlGetLeaders)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	leaders := make([]Leader, 0)
	for rows.Next() {

		var leader Leader
		rows.Scan(&leader.ID,
			&leader.Name,
			&leader.Image,
			&leader.Designation,
			&leader.Abbr,
			&leader.Featured,
			&leader.Description,
			&leader.CreatedAt,
			&leader.UpdatedAt)

		leaders = append(leaders, leader)
	}

	return leaders, nil
}
