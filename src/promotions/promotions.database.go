package promotions

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

func createPromotionInDb(promotion Promotion) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	featured, _ := strconv.ParseBool(*promotion.Featured)
	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO promotion(
															name,
															image,															
															label,
															price,
															featured,
															description
														)
														VALUES (
															?,?,?,?,?,?
														)`,
		promotion.Name,
		promotion.Image,
		promotion.Label,
		promotion.Price,
		featured,
		promotion.Description)

	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsInserted, _ := results.RowsAffected()
	status.SetStatus(numRowsInserted, 1)
	return status, nil
}

func deletePromotionsFromDb() (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status := &misc.Status{}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM promotion`)
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsDeleted, err := results.RowsAffected()
	status.SetStatus(numRowsDeleted, 1)
	return status, nil
}

func deletePromotionFromDb(promotionId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM promotion
														WHERE ID = ?`,
		promotionId)
	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsDeleted, err := results.RowsAffected()
	status.SetStatus(numRowsDeleted, 1)
	return status, nil
}

func buildUpdateSQLFromInput(promotionId int64, promotion Promotion) (string, []interface{}) {

	var sb strings.Builder
	args := make([]interface{}, 0)
	sb.WriteString("UPDATE promotion SET ")

	if promotion.Name != nil {
		sb.WriteString("name = ? ")
		args = append(args, *promotion.Name)
	}

	if promotion.Image != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("image = ? ")
		args = append(args, *promotion.Image)
	}

	if promotion.Label != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("label = ? ")
		args = append(args, *promotion.Label)
	}

	if promotion.Price != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("price = ? ")
		args = append(args, *promotion.Price)
	}

	if promotion.Featured != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		featured, _ := strconv.ParseBool(*promotion.Featured)
		sb.WriteString("featured = ? ")
		args = append(args, featured)
	}

	if promotion.Description != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("description = ? ")
		args = append(args, *promotion.Description)
	}

	// It seems we have nothing to update, just return
	// empty SQL and nil args
	if len(args) == 0 {
		return "", nil
	}

	sb.WriteString("WHERE id = ?")
	args = append(args, promotionId)
	return sb.String(), args
}

func updatePromotionFromDb(promotionId int64, promotion Promotion) (*Promotion, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateSql, updateArgs := buildUpdateSQLFromInput(promotionId, promotion)
	results, err := database.DbConn.ExecContext(ctx, updateSql, updateArgs...)
	if err != nil {
		log.Println("Error updating record ", promotionId)
		return nil, err
	}

	numRowsUpdated, _ := results.RowsAffected()
	if numRowsUpdated == 0 {
		return &Promotion{}, fmt.Errorf("No rows updated")
	}

	promotionUpdated, err := getPromotionFromDb(promotionId)
	return promotionUpdated, err
}

func getPromotionFromDb(promotionId int64) (*Promotion, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT
													id,
													name,
													image,
													label,
													price,
													CASE WHEN featured = 0 then 'false' ELSE 'true' END,
													description,
													createdAt,
													updatedAt
												FROM promotion
												WHERE id = ?`, promotionId)

	var promotion Promotion
	err := row.Scan(&promotion.ID,
		&promotion.Name,
		&promotion.Image,
		&promotion.Label,
		&promotion.Price,
		&promotion.Featured,
		&promotion.Description,
		&promotion.CreatedAt,
		&promotion.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &promotion, nil
}

func getPromotionsFromDb(isFeatured bool) ([]Promotion, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sqlGetPromotions := `SELECT 
							id,
							name,
							image,
							label,
							price,
							CASE WHEN featured = 0 then 'false' ELSE 'true' END,
							description,
							createdAt,
							updatedAt
						FROM promotion`

	if isFeatured {
		sqlGetPromotions += " WHERE featured = 1"
	}

	rows, err := database.DbConn.QueryContext(ctx, sqlGetPromotions)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	promotions := make([]Promotion, 0)
	for rows.Next() {

		var promotion Promotion
		rows.Scan(&promotion.ID,
			&promotion.Name,
			&promotion.Image,
			&promotion.Label,
			&promotion.Price,
			&promotion.Featured,
			&promotion.Description,
			&promotion.CreatedAt,
			&promotion.UpdatedAt)

		promotions = append(promotions, promotion)
	}

	return promotions, nil
}
