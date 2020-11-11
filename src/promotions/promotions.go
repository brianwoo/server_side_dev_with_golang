package promotions

type Promotion struct {
	ID          int64   `json:"_id"`
	Name        *string `json:"name"`
	Image       *string `json:"image"`
	Label       *string `json:"label"`
	Price       *string `json:"price"`
	Featured    *string `json:"featured"`
	Description *string `json:"description"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}
