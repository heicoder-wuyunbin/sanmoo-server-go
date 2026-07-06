package pagination

// PageData 统一分页响应结构（配合 response.Result.Data 使用）。
type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

func NewPageData(list any, total int64, page, size int) *PageData {
	q := Normalize(page, size)
	return &PageData{
		List:  list,
		Total: total,
		Page:  q.Page,
		Size:  q.Size,
	}
}
