package store

import (
	"fmt"
	"slices"
	"strings"
)

const (
	DBTypeSqlite   = "sqlite"
	DBTypePostgres = "postgres"
	DBTypeMySQL    = "mysql"
)

type FilterOperator string

const (
	OpEqual              FilterOperator = "="
	OpNotEqual           FilterOperator = "!="
	OpIn                 FilterOperator = "IN"
	OpGreaterThan        FilterOperator = ">"
	OpGreaterThanOrEqual FilterOperator = ">="
	OpLessThan           FilterOperator = "<"
	OpLessThanOrEqual    FilterOperator = "<="
	OpLike               FilterOperator = "LIKE"
)

const (
	BoolTrue  = "true"
	BoolFalse = "false"
)

type Filter struct {
	Field    string
	Operator FilterOperator
	Values   []string
}

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type ListQueryConditions struct {
	Filters    []Filter
	SearchTerm string
	SortField  string
	SortOrder  SortOrder
	Page       uint
	Limit      uint
}

type ListQueryBuilder struct {
	// Database type for dialect specific handling
	dbType string

	searchFields []string

	// Field name transformation (eg: "username" => "u.USERNAME")
	fieldTransformations map[string]string

	// Value transformations per field (eg: bool -> int for SQLite)
	valueTransformations map[string]func(v any) any

	// Default operator per field
	defaultOperators map[string]FilterOperator

	// Validation
	allowedFilterFields []string // NEW: Allowed fields for filtering
	allowedSortFields   []string // Allowed fields for sorting

	isBuilt bool
}

func NewQueryBuilder(dbType string) *ListQueryBuilder {
	if dbType == "" {
		dbType = DBTypeSqlite
	}
	return &ListQueryBuilder{
		dbType:               dbType,
		searchFields:         []string{},
		fieldTransformations: map[string]string{},
		valueTransformations: map[string]func(v any) any{},
		defaultOperators:     map[string]FilterOperator{},
		allowedFilterFields:  []string{},
		allowedSortFields:    []string{},
	}
}

func (b *ListQueryBuilder) WithSearchFields(fields ...string) *ListQueryBuilder {
	b.searchFields = fields
	return b
}

func (b *ListQueryBuilder) WithFieldTransformation(field, dbField string) *ListQueryBuilder {
	b.fieldTransformations[field] = dbField
	return b
}

func (b *ListQueryBuilder) WithBooleanField(field string) *ListQueryBuilder {
	if b.dbType == DBTypeSqlite {
		b.valueTransformations[field] = func(v any) any {
			if strVal, ok := v.(string); ok {
				if strVal == BoolTrue {
					return 1
				}
				return 0
			}
			if boolVal, ok := v.(bool); ok {
				if boolVal {
					return 1
				}
				return 0
			}
			return v
		}
	}
	return b
}

func (b *ListQueryBuilder) WithDefaultOperator(field string, op FilterOperator) *ListQueryBuilder {
	b.defaultOperators[field] = op
	return b
}

// WithAllowedFilterFields sets the allowed fields for filtering (validation)
func (b *ListQueryBuilder) WithAllowedFilterFields(fields ...string) *ListQueryBuilder {
	b.allowedFilterFields = fields
	return b
}

// WithAllowedSortFields sets the allowed fields for sorting (validation)
func (b *ListQueryBuilder) WithAllowedSortFields(fields ...string) *ListQueryBuilder {
	b.allowedSortFields = fields
	return b
}

func (b *ListQueryBuilder) Build(
	baseListQuery, countTotalQuery string,
	conditions *ListQueryConditions,
) (listQuery, totalQuery string, args []any, err error) {
	if b.isBuilt {
		return "", "", nil, fmt.Errorf("query builder already used, create a new instance")
	}
	b.isBuilt = true

	// Validate filter fields
	if len(b.allowedFilterFields) > 0 {
		for _, filter := range conditions.Filters {
			if !b.isFilterFieldAllowed(filter.Field) {
				return "", "", nil, fmt.Errorf("invalid filter field: %s", filter.Field)
			}
		}
	}

	// Validate sort field
	if conditions.SortField != "" && len(b.allowedSortFields) > 0 {
		if !b.isSortFieldAllowed(conditions.SortField) {
			return "", "", nil, fmt.Errorf("invalid sort field: %s", conditions.SortField)
		}
	}

	// Build WHERE clause
	whereClause, args, err := b.buildWhereClause(conditions)
	if err != nil {
		return "", "", nil, err
	}

	// Build list query
	var listBuilder strings.Builder
	listBuilder.WriteString(baseListQuery)

	// Add WHERE clause if we have conditions
	if whereClause != "" {
		if strings.Contains(baseListQuery, "WHERE") {
			listBuilder.WriteString(" AND ")
		} else {
			listBuilder.WriteString(" WHERE ")
		}
		listBuilder.WriteString(whereClause)
	}

	// Add ORDER BY
	if conditions.SortField != "" {
		listBuilder.WriteString(" ORDER BY ")
		listBuilder.WriteString(b.transformField(conditions.SortField))
		listBuilder.WriteString(" ")
		listBuilder.WriteString(string(conditions.SortOrder))
	}

	// Add pagination
	offset := (conditions.Page - 1) * conditions.Limit
	listBuilder.WriteString(" LIMIT ")
	listBuilder.WriteString(b.getPlaceholder(len(args) + 1))
	listBuilder.WriteString(" OFFSET ")
	listBuilder.WriteString(b.getPlaceholder(len(args) + 2))

	args = append(args, conditions.Limit, offset)
	listQuery = listBuilder.String()

	// Build count query (uses same WHERE clause but no pagination)
	totalQuery = countTotalQuery
	if whereClause != "" {
		if strings.Contains(countTotalQuery, "WHERE") {
			totalQuery += " AND " + whereClause
		} else {
			totalQuery += " WHERE " + whereClause
		}
	}

	return listQuery, totalQuery, args, nil
}

// buildWhereClause constructs the WHERE clause with filters and search
func (b *ListQueryBuilder) buildWhereClause(conditions *ListQueryConditions) (string, []any, error) {
	var whereClauses []string
	var args []any
	argIndex := 1

	// Process filters
	for _, filter := range conditions.Filters {
		dbField := b.transformField(filter.Field)

		switch filter.Operator {
		case OpIn:
			if len(filter.Values) == 0 {
				continue
			}
			placeholders := make([]string, len(filter.Values))
			for i, val := range filter.Values {
				transformedVal := b.transformValue(filter.Field, val)
				args = append(args, transformedVal)
				placeholders[i] = b.getPlaceholder(argIndex)
				argIndex++
			}
			whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", dbField, strings.Join(placeholders, ", ")))

		case OpEqual, OpNotEqual, OpGreaterThan, OpGreaterThanOrEqual, OpLessThan, OpLessThanOrEqual:
			if len(filter.Values) == 0 {
				continue
			}
			transformedVal := b.transformValue(filter.Field, filter.Values[0])
			whereClauses = append(whereClauses, fmt.Sprintf("%s %s %s", dbField, filter.Operator, b.getPlaceholder(argIndex)))
			args = append(args, transformedVal)
			argIndex++

		case OpLike:
			if len(filter.Values) == 0 {
				continue
			}
			whereClauses = append(whereClauses, fmt.Sprintf("%s LIKE %s", dbField, b.getPlaceholder(argIndex)))
			args = append(args, fmt.Sprintf("%%%v%%", filter.Values[0]))
			argIndex++
		}
	}

	// Process search term
	if conditions.SearchTerm != "" && len(b.searchFields) > 0 {
		searchClauses := make([]string, len(b.searchFields))
		for i, field := range b.searchFields {
			dbField := b.transformField(field)
			searchClauses[i] = fmt.Sprintf("%s LIKE %s", dbField, b.getPlaceholder(argIndex))
			args = append(args, "%"+conditions.SearchTerm+"%")
			argIndex++
		}
		whereClauses = append(whereClauses, "("+strings.Join(searchClauses, " OR ")+")")
	}

	if len(whereClauses) == 0 {
		return "", args, nil
	}

	return strings.Join(whereClauses, " AND "), args, nil
}

// isFilterFieldAllowed checks if a filter field is in the allowed list
func (b *ListQueryBuilder) isFilterFieldAllowed(field string) bool {
	// Transform field first (e.g., "username" -> "u.USERNAME")
	transformedField := b.transformField(field)

	// Check both original and transformed field
	return slices.Contains(b.allowedFilterFields, field) ||
		slices.Contains(b.allowedFilterFields, transformedField)
}

// isSortFieldAllowed checks if a sort field is in the allowed list
func (b *ListQueryBuilder) isSortFieldAllowed(field string) bool {
	// Transform field first
	transformedField := b.transformField(field)

	// Check both original and transformed field
	return slices.Contains(b.allowedSortFields, field) ||
		slices.Contains(b.allowedSortFields, transformedField)
}

// transformField converts field if transform is configured
func (b *ListQueryBuilder) transformField(field string) string {
	if dbField, ok := b.fieldTransformations[field]; ok {
		return dbField
	}
	return field
}

// transformValue converts value if transform is configured
func (b *ListQueryBuilder) transformValue(field string, value any) any {
	if transform, ok := b.valueTransformations[field]; ok {
		return transform(value)
	}
	return value
}

// getPlaceholder returns the appropriate placeholder for the database type
func (b *ListQueryBuilder) getPlaceholder(index int) string {
	switch b.dbType {
	case DBTypePostgres:
		return fmt.Sprintf("$%d", index)
	case DBTypeSqlite, DBTypeMySQL:
		return "?"
	default:
		return "?"
	}
}
