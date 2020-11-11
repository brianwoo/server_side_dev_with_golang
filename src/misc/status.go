package misc

// Status mimic Mongoose's Update, Delete status
type Status struct {
	NumOfRowsAffected int64 `json:"n"`
	IsOk              int8  `json:"ok"`
}

func (ds *Status) SetStatus(numOfRowsAffected int64, isOk int8) {
	ds.NumOfRowsAffected = numOfRowsAffected
	ds.IsOk = isOk
}
