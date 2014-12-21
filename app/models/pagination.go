package models

// Using a paginator will allow queries
// to know with confidence that they can
// be sure that the pagination is handled correctly.
// Being an interface allows each type to set requirements
// for what is allowed on limits and sort (fields, etc)
type Paginator interface {
	Offset() int64
	Limit() int64
	Sort() string
}

type DefaultPagination struct {
	Page      int64
	PerPage   int64
	SortOrder string
	SortKey   string
}

func (p *DefaultPagination) ApplyDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}

	if p.PerPage == 0 {
		p.PerPage = 25
	}

	if p.SortOrder == "" {
		p.SortOrder = "DESC"
	}

	if p.SortKey == "" {
		p.SortKey = "id"
	}
}

func (p DefaultPagination) Offset() int64 {
	return (p.Page - 1) * p.PerPage
}

func (p DefaultPagination) Limit() int64 {
	return p.PerPage
}

func (p DefaultPagination) Sort() string {
	return p.SortKey + " " + p.SortOrder
}
