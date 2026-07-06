package pagination

// PageQuery 表示分页参数。
type PageQuery struct {
	Page int
	Size int
}

func Normalize(page, size int) PageQuery {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return PageQuery{Page: page, Size: size}
}

func Offset(page, size int) int {
	return (page - 1) * size
}
