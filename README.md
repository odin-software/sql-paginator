# SQL Paginator

A type-safe, easy-to-use SQL pagination library for Go that helps you implement paginated database queries with minimal boilerplate.

## Features

- 🔒 Type-safe pagination using Go generics
- 📦 Simple API for paginated database queries
- 🚀 Works with standard `database/sql`
- 🛠️ Built-in HTTP request parameter parsing
- 📊 Complete pagination metadata (total items, total pages, etc.)
- ✅ Fully tested with comprehensive test suite

## Installation

```bash
go get github.com/odin-software/sql-paginator
```

## Quick Start

```go
// Define your model
type User struct {
    ID   int
    Name string
}

// Create a scan function for your model
func scanUser(rows *sql.Rows) (User, error) {
    var u User
    err := rows.Scan(&u.ID, &u.Name)
    return u, err
}

// Initialize the paginator
db, err := sql.Open("postgres", "postgres://localhost/mydb?sslmode=disable")
paginator := NewPaginator[User](db)

// Use the paginator
result, err := paginator.QueryPaginated(
    context.Background(),
    "SELECT id, name FROM users WHERE active = true",
    []any{},  // Query arguments (if any)
    1,        // Page number
    10,       // Items per page
    scanUser, // Scan function
)

if err != nil {
    log.Fatal(err)
}

// Access paginated results
fmt.Printf("Total items: %d\n", result.Total)
fmt.Printf("Total pages: %d\n", result.TotalPages)
fmt.Printf("Current page: %d\n", result.Page)
fmt.Printf("Items per page: %d\n", result.Limit)

for _, user := range result.Items {
    fmt.Printf("User: %s (ID: %d)\n", user.Name, user.ID)
}
```

## HTTP Handler Example

```go
func UsersHandler(w http.ResponseWriter, r *http.Request) {
    // Get page and limit from query parameters
    page, limit, err := GetPageAndLimitParams(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    paginator := NewPaginator[User](db)
    result, err := paginator.QueryPaginated(
        r.Context(),
        "SELECT id, name FROM users",
        nil,
        page,
        limit,
        scanUser,
    )

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}
```

## Pagination Response Structure

The paginator returns a `Page[T]` struct containing:

```go
type Page[T any] struct {
    Items      []T `json:"items"`       // Slice of items for current page
    Page       int `json:"page"`        // Current page number
    Limit      int `json:"limit"`       // Items per page
    Total      int `json:"total"`       // Total number of items
    TotalPages int `json:"total_pages"` // Total number of pages
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
