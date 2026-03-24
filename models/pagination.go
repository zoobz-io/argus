package models

// DefaultPageSize is the default number of results per page.
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size.
const MaxPageSize = 100

// CursorPage holds cursor-based pagination parameters for store queries.
type CursorPage struct {
	Cursor *int64
	Limit  int
}

// PageSize returns the effective page size, clamped to bounds.
func (p CursorPage) PageSize() int {
	if p.Limit <= 0 {
		return DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		return MaxPageSize
	}
	return p.Limit
}

// CursorResult holds cursor-based pagination metadata alongside results.
type CursorResult[T any] struct {
	Cursor  *int64
	Items   []*T
	HasMore bool
}
