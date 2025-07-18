package paginator

import (
	"context"
	"database/sql"
	"fmt"
)

type Paginator[T any] struct {
	DB *sql.DB
}

type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginator[T any](db *sql.DB) *Paginator[T] {
	return &Paginator[T]{
		DB: db,
	}
}

func (p *Paginator[T]) QueryPaginated(
	context context.Context,
	query string,
	args []any,
	page int,
	limit int,
	scan func(*sql.Rows) (T, error),
) (*Page[T], error) {
	offset := (page - 1) * limit

	limitClause := fmt.Sprintf(" OFFSET $%d LIMIT $%d", len(args)+1, len(args)+2)

	formattedQuery := query + limitClause

	argsWithLimit := append(args, offset, limit)

	rows, err := p.DB.QueryContext(context, formattedQuery, argsWithLimit...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]T, 0)

	for rows.Next() {
		item, err := scan(rows)

		if err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	totalQuery := `
		SELECT COUNT(*) FROM (
			` + query + `
		) AS subquery
	`

	var totalRows int

	err = p.DB.QueryRowContext(context, totalQuery, args...).Scan(&totalRows)

	if err != nil {
		return nil, err
	}

	totalPages := (totalRows + limit - 1) / limit

	return &Page[T]{
		Items:      results,
		Page:       page,
		Limit:      limit,
		Total:      totalRows,
		TotalPages: totalPages,
	}, rows.Err()
}
