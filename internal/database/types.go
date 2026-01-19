package database

// PaginatedResponse представляет ответ с пагинацией
type PaginatedResponse[T any] struct {
	Data            []T  `json:"data"`
	Total           int  `json:"total"`
	Limit           int  `json:"limit"`
	Offset          int  `json:"offset"`
	HasNextPage     bool `json:"has_next_page"`
	HasPreviousPage bool `json:"has_previous_page"`
}

// NewPaginatedResponse создает новый PaginatedResponse
func NewPaginatedResponse[T any](data []T, total, limit, offset int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:            data,
		Total:           total,
		Limit:           limit,
		Offset:          offset,
		HasNextPage:     offset+limit < total,
		HasPreviousPage: offset > 0,
	}
}
