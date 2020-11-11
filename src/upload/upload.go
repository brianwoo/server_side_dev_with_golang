package upload

type UploadResult struct {
	FieldName    string `json:"fieldname"`
	OriginalName string `json:"originalname"`
	Encoding     string `json:"encoding"`
	MimeType     string `json:"mimetype"`
	Destination  string `json:"destination"`
	Filename     string `json:"filename"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
}
