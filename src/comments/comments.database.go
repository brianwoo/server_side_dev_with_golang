package comments

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"confusion.com/bwoo/database"
	"confusion.com/bwoo/misc"
)

func createCommentInDb(dishId int64, authorId int64, comment Comment) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO comment (
															dishId,
															rating,
															comment,
															authorId
														) 
														VALUES (
															?,?,?,?
														)`,
		dishId,
		comment.Rating,
		comment.Comment,
		authorId)

	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsInserted, _ := results.RowsAffected()
	status.SetStatus(numRowsInserted, 1)
	return status, nil
}

func deleteCommentsFromDb(dishId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status := &misc.Status{}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM comment WHERE dishId = ?`, dishId)
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsDeleted, err := results.RowsAffected()
	status.SetStatus(numRowsDeleted, 1)
	return status, nil
}

func deleteCommentFromDb(dishId, commentId, updatedByUserId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM comment
														WHERE dishid = ? AND id = ? and authorId = ?`,
		dishId, commentId, updatedByUserId)
	commentStatus := &misc.Status{}
	if err != nil {
		commentStatus.SetStatus(0, 0)
		return commentStatus, err
	}

	numRowsDeleted, err := results.RowsAffected()
	commentStatus.SetStatus(numRowsDeleted, 1)
	return commentStatus, nil
}

/**
* We only allow the user update the Rating or the Comment
 */
func buildUpdateSQLFromInput(dishId int64, commentId int64, comment Comment, updatedByUserId int64) (string, []interface{}) {

	var sb strings.Builder
	args := make([]interface{}, 0)
	sb.WriteString("UPDATE comment SET ")

	if comment.Rating != nil {
		sb.WriteString("rating = ? ")
		args = append(args, *comment.Rating)
	}

	if comment.Comment != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("comment = ? ")
		args = append(args, *comment.Comment)
	}

	// It seems we have nothing to update, just return
	// empty SQL and nil args
	if len(args) == 0 {
		return "", nil
	}

	sb.WriteString("WHERE dishId = ? AND id = ? AND authorId = ?")
	args = append(args, dishId)
	args = append(args, commentId)
	args = append(args, updatedByUserId)
	return sb.String(), args
}

func updateCommentFromDb(dishId int64, commentId int64, comment Comment, updatedByUserId int64) (*Comment, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateSql, updateArgs := buildUpdateSQLFromInput(dishId, commentId, comment, updatedByUserId)
	results, err := database.DbConn.ExecContext(ctx, updateSql, updateArgs...)
	if err != nil {
		log.Println("Error updating record ", dishId)
		return nil, err
	}

	numRowsUpdated, _ := results.RowsAffected()
	if numRowsUpdated == 0 {
		return &Comment{}, fmt.Errorf("No rows updated")
	}

	commentUpdated, err := getCommentFromDb(dishId, commentId)
	return commentUpdated, err
}

func getCommentFromDb(dishId, commentId int64) (*Comment, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT
													c.Id,
													c.rating,
													c.comment,
													u.firstname,
													u.lastname,
													u.Id,
													c.date
												FROM comment c, user u
												WHERE c.authorId = u.id
												AND c.dishId = ? 
												AND c.id = ?`,
		dishId, commentId)

	var comment Comment
	var author Author
	comment.Author = &author
	err := row.Scan(&comment.ID,
		&comment.Rating,
		&comment.Comment,
		&comment.Author.Firstname,
		&comment.Author.Lastname,
		&comment.Author.ID,
		&comment.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &comment, nil
}

func getCommentsFromDb(dishId int64) ([]Comment, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := database.DbConn.QueryContext(ctx, `SELECT 
														c.id,
														c.rating,
														c.comment,
														u.firstname,
														u.lastname,
														c.date
													FROM comment c, user u
													WHERE c.authorId = u.id
													AND c.dishId = ?`, dishId)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	comments := make([]Comment, 0)
	for rows.Next() {

		var comment Comment
		var author Author
		comment.Author = &author
		rows.Scan(&comment.ID,
			&comment.Rating,
			&comment.Comment,
			&comment.Author.Firstname,
			&comment.Author.Lastname,
			&comment.Date)

		comments = append(comments, comment)
	}

	return comments, nil
}
