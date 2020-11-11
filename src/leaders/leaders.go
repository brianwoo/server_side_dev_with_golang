package leaders

type Leader struct {
	ID          int64   `json:"_id"`
	Name        *string `json:"name"`
	Image       *string `json:"image"`
	Designation *string `json:"designation"`
	Abbr        *string `json:"abbr"`
	Featured    *string `json:"featured"`
	Description *string `json:"description"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}
