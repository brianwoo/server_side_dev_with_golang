package comments

type Author struct {
	ID        int64  `json:"_id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type Comment struct {
	ID      int64   `json:"_id"`
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
	Author  *Author `json:"author"`
	Date    *string `json:"date"`
}
