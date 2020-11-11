package favoriteDishes

import "confusion.com/bwoo/dishes"

type favoriteDish struct {
	ID int64 `json:"_id"`
}

type favoriteDishExist struct {
	IsExists  bool         `json:"exists"`
	Favorites *dishes.Dish `json:"favorites"`
}

type favoriteDishesResult struct {
	Dishes []dishes.Dish `json:"dishes"`
}

type favoriteDishes []favoriteDish
