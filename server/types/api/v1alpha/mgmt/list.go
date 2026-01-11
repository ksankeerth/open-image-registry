package mgmt

type EntityListResponse[T any] struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	Limit    int `json:"limit"`
	Entities []T `json:"entities"`
}