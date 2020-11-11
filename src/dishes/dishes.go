package dishes

import (
	"confusion.com/bwoo/comments"
)

type Dish struct {
	ID          int64              `json:"_id"`
	Name        *string            `json:"name"`
	Image       *string            `json:"image"`
	Category    *string            `json:"category"`
	Label       *string            `json:"label"`
	Price       *string            `json:"price"`
	Featured    *string            `json:"featured"`
	Description *string            `json:"description"`
	Comments    []comments.Comment `json:"comments"`
	CreatedAt   *string            `json:"createdAt"`
	UpdatedAt   *string            `json:"updatedAt"`
}
