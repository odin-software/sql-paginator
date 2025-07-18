package paginator

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type User struct {
	ID   int
	Name string
}

func scanUser(rows *sql.Rows) (User, error) {
	var u User
	err := rows.Scan(&u.ID, &u.Name)
	return u, err
}

func TestQueryPaginatedWithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	page := 1
	limit := 2
	offset := 0

	baseQuery := "SELECT id, name FROM users WHERE active = true"
	expectedQuery := regexp.QuoteMeta(baseQuery + " OFFSET $1 LIMIT $2")

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery(expectedQuery).
		WithArgs(offset, limit).
		WillReturnRows(rows)

	countQueryPattern := `SELECT COUNT\(\*\) FROM \( ` + regexp.QuoteMeta(baseQuery) + ` \) AS subquery`
	mock.ExpectQuery(countQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	p := NewPaginator[User](db)

	result, err := p.QueryPaginated(context.Background(), baseQuery, []any{}, page, limit, scanUser)
	require.NoError(t, err)

	require.Equal(t, 2, len(result.Items))
	require.Equal(t, 5, result.Total)
	require.Equal(t, 3, result.TotalPages)
	require.Equal(t, "Alice", result.Items[0].Name)
	require.Equal(t, "Bob", result.Items[1].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryPaginatedReturnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	page := 1
	limit := 1
	offset := 0

	baseQuery := "SELECT id, name FROM users"
	expectedQuery := regexp.QuoteMeta(baseQuery + " OFFSET $1 LIMIT $2")

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("not-an-int", "Charlie")

	mock.ExpectQuery(expectedQuery).
		WithArgs(offset, limit).
		WillReturnRows(rows)

	p := NewPaginator[User](db)

	_, err = p.QueryPaginated(context.Background(), baseQuery, []any{}, page, limit, scanUser)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Scan error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryPaginatedWithMultiplePages(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	page := 2
	limit := 2
	offset := 2

	baseQuery := "SELECT id, name FROM users"
	expectedQuery := regexp.QuoteMeta(baseQuery + " OFFSET $1 LIMIT $2")

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(3, "Charlie").
		AddRow(4, "David")

	mock.ExpectQuery(expectedQuery).
		WithArgs(offset, limit).
		WillReturnRows(rows)

	countQueryPattern := `SELECT COUNT\(\*\) FROM \( ` + regexp.QuoteMeta(baseQuery) + ` \) AS subquery`
	mock.ExpectQuery(countQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(6))

	p := NewPaginator[User](db)

	result, err := p.QueryPaginated(context.Background(), baseQuery, []any{}, page, limit, scanUser)
	require.NoError(t, err)

	require.Equal(t, 2, len(result.Items))
	require.Equal(t, 6, result.Total)
	require.Equal(t, 3, result.TotalPages)
	require.Equal(t, 2, result.Page)
	require.Equal(t, "Charlie", result.Items[0].Name)
	require.Equal(t, "David", result.Items[1].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryPaginatedWithNoResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	page := 1
	limit := 10
	offset := 0

	baseQuery := "SELECT id, name FROM users WHERE 1 = 0"
	expectedQuery := regexp.QuoteMeta(baseQuery + " OFFSET $1 LIMIT $2")

	rows := sqlmock.NewRows([]string{"id", "name"})

	mock.ExpectQuery(expectedQuery).
		WithArgs(offset, limit).
		WillReturnRows(rows)

	countQueryPattern := `SELECT COUNT\(\*\) FROM \( ` + regexp.QuoteMeta(baseQuery) + ` \) AS subquery`
	mock.ExpectQuery(countQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	p := NewPaginator[User](db)

	result, err := p.QueryPaginated(context.Background(), baseQuery, []any{}, page, limit, scanUser)
	require.NoError(t, err)

	require.Equal(t, 0, len(result.Items))
	require.Equal(t, 0, result.Total)
	require.Equal(t, 0, result.TotalPages)
	require.Equal(t, 1, result.Page)
	require.Equal(t, 10, result.Limit)

	require.NoError(t, mock.ExpectationsWereMet())
}
