package file

type FileItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	URL        string `json:"url"`
	CreateTime string `json:"createTime"`
}
