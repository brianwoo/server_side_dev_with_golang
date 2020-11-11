package dishes

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

func createDishInDb(dish Dish) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	featured, _ := strconv.ParseBool(*dish.Featured)
	results, err := database.DbConn.ExecContext(ctx, `INSERT INTO dish(
															name,
															image,
															category,
															label,
															price,
															featured,
															description
														)
														VALUES (
															?,?,?,?,?,?,?
														)`,
		dish.Name,
		dish.Image,
		dish.Category,
		dish.Label,
		dish.Price,
		featured,
		dish.Description)

	status := &misc.Status{}
	if err != nil {
		status.SetStatus(0, 0)
		return status, err
	}

	numRowsInserted, _ := results.RowsAffected()
	status.SetStatus(numRowsInserted, 1)
	return status, nil
}

func deleteDishesFromDb() (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dishStatus := &misc.Status{}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM dish`)
	if err != nil {
		dishStatus.SetStatus(0, 0)
		return dishStatus, err
	}

	numRowsDeleted, err := results.RowsAffected()
	dishStatus.SetStatus(numRowsDeleted, 1)
	return dishStatus, nil
}

func deleteDishFromDb(dishId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM dish
														WHERE ID = ?`,
		dishId)
	dishStatus := &misc.Status{}
	if err != nil {
		dishStatus.SetStatus(0, 0)
		return dishStatus, err
	}

	numRowsDeleted, err := results.RowsAffected()
	dishStatus.SetStatus(numRowsDeleted, 1)
	return dishStatus, nil
}

func buildUpdateSQLFromInput(dishId int64, dish Dish) (string, []interface{}) {

	var sb strings.Builder
	args := make([]interface{}, 0)
	sb.WriteString("UPDATE dish SET ")

	if dish.Name != nil {
		sb.WriteString("name = ? ")
		args = append(args, *dish.Name)
	}

	if dish.Image != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("image = ? ")
		args = append(args, *dish.Image)
	}

	if dish.Category != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("category = ? ")
		args = append(args, *dish.Category)
	}

	if dish.Label != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("label = ? ")
		args = append(args, *dish.Label)
	}

	if dish.Price != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("price = ? ")
		args = append(args, *dish.Price)
	}

	if dish.Featured != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		featured, _ := strconv.ParseBool(*dish.Featured)
		sb.WriteString("featured = ? ")
		args = append(args, featured)
	}

	if dish.Description != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("description = ? ")
		args = append(args, *dish.Description)
	}

	// It seems we have nothing to update, just return
	// empty SQL and nil args
	if len(args) == 0 {
		return "", nil
	}

	sb.WriteString("WHERE id = ?")
	args = append(args, dishId)
	return sb.String(), args
}

func updateDishFromDb(dishId int64, dish Dish) (*Dish, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateSql, updateArgs := buildUpdateSQLFromInput(dishId, dish)
	results, err := database.DbConn.ExecContext(ctx, updateSql, updateArgs...)
	if err != nil {
		log.Println("Error updating record ", dishId)
		return nil, err
	}

	numRowsUpdated, _ := results.RowsAffected()
	if numRowsUpdated == 0 {
		return &Dish{}, fmt.Errorf("No rows updated")
	}

	dishUpdated, err := getDishFromDb(dishId)
	return dishUpdated, err
}

func getDishFromDb(dishId int64) (*Dish, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT 
													id,
													name,
													image,
													category,
													label,
													price,
													CASE WHEN featured = 0 then 'false' ELSE 'true' END,
													description,
													createdAt,
													updatedAt
												FROM dish
												WHERE id = ?`, dishId)

	var dish Dish
	err := row.Scan(&dish.ID,
		&dish.Name,
		&dish.Image,
		&dish.Category,
		&dish.Label,
		&dish.Price,
		&dish.Featured,
		&dish.Description,
		&dish.CreatedAt,
		&dish.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &dish, nil
}

func getDishesFromDb(isFeatured bool) ([]Dish, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sqlGetDishes := `SELECT 
						id,
						name,
						image,
						category,
						label,
						price,
						CASE WHEN featured = 0 then 'false' ELSE 'true' END,
						description,
						createdAt,
						updatedAt
					FROM dish`

	if isFeatured {
		sqlGetDishes += " WHERE featured = 1"
	}

	rows, err := database.DbConn.QueryContext(ctx, sqlGetDishes)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	dishes := make([]Dish, 0)
	for rows.Next() {

		var dish Dish
		rows.Scan(&dish.ID,
			&dish.Name,
			&dish.Image,
			&dish.Category,
			&dish.Label,
			&dish.Price,
			&dish.Featured,
			&dish.Description,
			&dish.CreatedAt,
			&dish.UpdatedAt)

		dishes = append(dishes, dish)
	}

	return dishes, nil
}
