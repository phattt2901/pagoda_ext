package pager

import (
	"fmt"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/tests"

	"github.com/stretchr/testify/assert"
)

func TestNewPager(t *testing.T) {
	e := echo.New()
	ctx, _ := tests.NewContext(e, "/")
	pgr := NewPager(ctx, 10)
	assert.Equal(t, 10, pgr.ItemsPerPage)
	assert.Equal(t, 1, pgr.Page)
	assert.Equal(t, 0, pgr.Items)
	assert.Equal(t, 1, pgr.Pages)

	ctx, _ = tests.NewContext(e, fmt.Sprintf("/abc?%s=%d", QueryKey, 2))
	pgr = NewPager(ctx, 10)
	assert.Equal(t, 2, pgr.Page)

	ctx, _ = tests.NewContext(e, fmt.Sprintf("/abc?%s=%d", QueryKey, -2))
	pgr = NewPager(ctx, 10)
	assert.Equal(t, 1, pgr.Page)
}

func TestPager_SetItems(t *testing.T) {
	ctx, _ := tests.NewContext(echo.New(), "/")
	pgr := NewPager(ctx, 20)
	pgr.SetItems(100)
	assert.Equal(t, 100, pgr.Items)
	assert.Equal(t, 5, pgr.Pages)

	pgr.SetItems(0)
	assert.Equal(t, 0, pgr.Items)
	assert.Equal(t, 1, pgr.Pages)
	assert.Equal(t, 1, pgr.Page)
}

func TestPager_IsBeginning(t *testing.T) {
	ctx, _ := tests.NewContext(echo.New(), "/")
	pgr := NewPager(ctx, 20)
	pgr.Pages = 10
	assert.True(t, pgr.IsBeginning())
	pgr.Page = 2
	assert.False(t, pgr.IsBeginning())
	pgr.Page = 1
	assert.True(t, pgr.IsBeginning())
}

func TestPager_IsEnd(t *testing.T) {
	ctx, _ := tests.NewContext(echo.New(), "/")
	pgr := NewPager(ctx, 20)
	pgr.Pages = 10
	assert.False(t, pgr.IsEnd())
	pgr.Page = 10
	assert.True(t, pgr.IsEnd())
	pgr.Page = 1
	assert.False(t, pgr.IsEnd())
}

func TestPager_GetOffset(t *testing.T) {
	ctx, _ := tests.NewContext(echo.New(), "/")
	pgr := NewPager(ctx, 20)
	assert.Equal(t, 0, pgr.GetOffset())
	pgr.Page = 2
	assert.Equal(t, 20, pgr.GetOffset())
	pgr.Page = 3
	assert.Equal(t, 40, pgr.GetOffset())
}

func TestPager_PageURL(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name               string
		initialURL         string // Full path with query, e.g., "/items?filter=active"
		itemsPerPage       int
		targetPage         int
		expectedQueryPairs map[string]string // Expected key-value pairs, order doesn't matter
	}{
		{
			name:               "no params, target page 2",
			initialURL:         "/items",
			itemsPerPage:       10,
			targetPage:         2,
			expectedQueryPairs: map[string]string{"page": "2"},
		},
		{
			name:               "no params, target page 1",
			initialURL:         "/items",
			itemsPerPage:       10,
			targetPage:         1,
			expectedQueryPairs: map[string]string{}, // Page 1, page param omitted
		},
		{
			name:               "no params, target page 0 (should be like 1)",
			initialURL:         "/items",
			itemsPerPage:       10,
			targetPage:         0,
			expectedQueryPairs: map[string]string{}, // Page 1, page param omitted
		},
		{
			name:               "existing param, target page 2",
			initialURL:         "/items?filter=active",
			itemsPerPage:       10,
			targetPage:         2,
			expectedQueryPairs: map[string]string{"filter": "active", "page": "2"},
		},
		{
			name:               "existing param, target page 1",
			initialURL:         "/items?filter=active",
			itemsPerPage:       10,
			targetPage:         1,
			expectedQueryPairs: map[string]string{"filter": "active"}, // Page 1, page param omitted
		},
		{
			name:               "multiple existing params, target page 3",
			initialURL:         "/items?filter=active&sort=name",
			itemsPerPage:       10,
			targetPage:         3,
			expectedQueryPairs: map[string]string{"filter": "active", "sort": "name", "page": "3"},
		},
		{
			name:               "existing page param, target new page",
			initialURL:         "/items?page=2&filter=pending",
			itemsPerPage:       10,
			targetPage:         3,
			expectedQueryPairs: map[string]string{"filter": "pending", "page": "3"},
		},
		{
			name:               "existing page param, target page 1",
			initialURL:         "/items?page=2&filter=pending",
			itemsPerPage:       10,
			targetPage:         1,
			expectedQueryPairs: map[string]string{"filter": "pending"},
		},
		{
			name:               "param with special chars",
			initialURL:         "/items?search=hello%20world%20%26%20co&existing=true", // Pre-encoded
			itemsPerPage:       10,
			targetPage:         2,
			expectedQueryPairs: map[string]string{"search": "hello world & co", "existing": "true", "page": "2"},
		},
		{
			name:               "empty initial query, target page 1",
			initialURL:         "/items?",
			itemsPerPage:       10,
			targetPage:         1,
			expectedQueryPairs: map[string]string{},
		},
		{
			name:               "empty initial query, target page 2",
			initialURL:         "/items?",
			itemsPerPage:       10,
			targetPage:         2,
			expectedQueryPairs: map[string]string{"page": "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Using NewContext from pkg/tests which sets up a basic echo.Context
			ctx, _ := tests.NewContext(e, tt.initialURL)
			p := NewPager(ctx, tt.itemsPerPage)
			actualURLString := p.PageURL(tt.targetPage)

			if len(tt.expectedQueryPairs) == 0 {
				assert.Equal(t, "", actualURLString, "URL string should be empty if no params")
				return
			}

			// Parse the actual URL string
			parsedActualURL, err := url.Parse(actualURLString)
			assert.NoError(t, err, "Generated URL string should be parseable")
			actualQueryParams := parsedActualURL.Query()

			// Compare lengths
			assert.Equal(t, len(tt.expectedQueryPairs), len(actualQueryParams), "Number of query parameters should match")

			// Compare key-value pairs
			for k, expectedVal := range tt.expectedQueryPairs {
				actualVal, ok := actualQueryParams[k]
				assert.True(t, ok, fmt.Sprintf("Expected query parameter '%s' to be present", k))
				if ok {
					assert.Equal(t, expectedVal, actualVal[0], fmt.Sprintf("Value for query parameter '%s' should match", k))
				}
			}
		})
	}
}
