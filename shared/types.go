package shared

type Pagination struct {
	Skip  int `form:"skip" binding:"min=0"`
	Limit int `form:"limit" binding:"min=0,max=50"`
}

type ListPage[T interface{}] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

// TF is this
type ErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

type TranslationText struct {
	En string `json:"en"`
	Th string `json:"th"`
}
