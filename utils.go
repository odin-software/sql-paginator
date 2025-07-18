package paginator

import (
	"net/http"
	"strconv"
)

func GetPageAndLimitParams(r *http.Request) (int, int) {
	page := 1
	limit := 10

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	return page, limit
}
