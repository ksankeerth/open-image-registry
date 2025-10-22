package query

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type SortCondition struct {
	Field string
	Order SortOrder
}

type Pagination struct {
	Page  uint
	Limit uint
}

type FilterCondition struct {
	Field  string
	Values []any
}

type ListModelsConditions struct {
	Filters    []FilterCondition
	Sort       *SortCondition
	Pagination Pagination
	SearchTerm string
}