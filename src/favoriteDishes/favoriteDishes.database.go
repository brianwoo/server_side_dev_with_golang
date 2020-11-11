package favoriteDishes

import (
	"context"
	"database/sql"
	"time"

	"confusion.com/bwoo/dishes"

	"confusion.com/bwoo/database"
	"confusion.com/bwoo/misc"
)

func getFavoriteDishFromDb(userId, dishId int64) (favoriteDishExist, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `SELECT 
													d.id,
													d.name,
													d.image,
													d.category,
													d.label,
													d.price,
													CASE WHEN d.featured = 0 then 'false' ELSE 'true' END,
													d.description,
													d.createdAt,
													d.updatedAt 
												FROM dish d, favoriteDish fd 
												WHERE d.id = fd.dishId 
												AND fd.userId = ? 
												AND fd.dishId = ?`,
		userId,
		dishId)

	var favDishExist favoriteDishExist
	var favDish dishes.Dish
	err := row.Scan(&favDish.ID,
		&favDish.Name,
		&favDish.Image,
		&favDish.Category,
		&favDish.Label,
		&favDish.Price,
		&favDish.Featured,
		&favDish.Description,
		&favDish.CreatedAt,
		&favDish.UpdatedAt)

	if err == sql.ErrNoRows {
		return favDishExist, nil
	} else if err != nil {
		return favDishExist, err
	}

	favDishExist.IsExists = true
	favDishExist.Favorites = &favDish
	return favDishExist, nil
}

func getFavoriteDishesFromDb(userId int64) (favoriteDishesResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := database.DbConn.QueryContext(ctx, `SELECT 
														d.id,
														d.name,
														d.image,
														d.category,
														d.label,
														d.price,
														CASE WHEN d.featured = 0 then 'false' ELSE 'true' END,
														d.description,
														d.createdAt,
														d.updatedAt 
														FROM dish d, favoriteDish fd 
														WHERE d.id = fd.dishId 
														AND fd.userId = ? `,
		userId)

	defer rows.Close()
	if err != nil {
		return favoriteDishesResult{}, err
	}

	favDishes := make([]dishes.Dish, 0)
	for rows.Next() {

		var dish dishes.Dish
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

		favDishes = append(favDishes, dish)
	}

	var favDishesResult favoriteDishesResult
	favDishesResult.Dishes = favDishes
	return favDishesResult, nil
}

func getExecContextFunc(tx *sql.Tx) func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

	if tx != nil {
		return func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			return tx.ExecContext(ctx, query, args...)
		}
	} else {
		return func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			return database.DbConn.ExecContext(ctx, query, args...)
		}
	}
}

func createFavoriteDishInDb(userId, dishId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return createFavoriteDishInDbInternal(nil, ctx, userId, dishId)
}

func createFavoriteDishInDbInternal(tx *sql.Tx, ctx context.Context, userId, dishId int64) (*misc.Status, error) {

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	sqlInsert := `INSERT INTO favoriteDish(
					userId,
					dishId
				)
				VALUES (
					?,?
				)`

	execContextFunc := getExecContextFunc(tx)
	result, err := execContextFunc(ctx, sqlInsert, userId, dishId)
	if err != nil {
		return status, err
	}

	rowsInserted, _ := result.RowsAffected()
	status.NumOfRowsAffected = rowsInserted
	status.IsOk = 1
	return status, nil
}

func createFavoriteDishesInDb(userId int64, favDishes favoriteDishes) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// we use transaction so we can commit atomicly (i.e. all or none)
	tx, err := database.DbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	for _, favDish := range favDishes {

		result, err := createFavoriteDishInDbInternal(tx, ctx, userId, favDish.ID)
		if err != nil {
			tx.Rollback()
			status.NumOfRowsAffected = 0
			return status, err
		}

		status.NumOfRowsAffected += result.NumOfRowsAffected
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	status.IsOk = 1
	return status, nil
}

func deleteFavoriteDishesFromDb(userId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM favoriteDish 
														WHERE userId = ?`,
		userId)

	if err != nil {
		return status, err
	}

	numOfRowsDeleted, _ := results.RowsAffected()
	status.NumOfRowsAffected = numOfRowsDeleted
	status.IsOk = 1
	return status, nil
}

func deleteFavoriteDishFromDb(userId, dishId int64) (*misc.Status, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status := &misc.Status{NumOfRowsAffected: 0, IsOk: 0}
	results, err := database.DbConn.ExecContext(ctx, `DELETE FROM favoriteDish 
														WHERE userId = ?
														AND dishId = ?`,
		userId,
		dishId)

	if err != nil {
		return status, err
	}

	numOfRowsDeleted, _ := results.RowsAffected()
	status.NumOfRowsAffected = numOfRowsDeleted
	status.IsOk = 1
	return status, nil
}
