package pager

import (
	"math"
	"net/url" // Added for url.Values
	"strconv"

	"github.com/labstack/echo/v4"
)

// QueryKey stores the query key used to indicate the current page.
const QueryKey = "page"

// Pager provides a mechanism to allow a user to page results via a query parameter.
type Pager struct {
	// Items stores the total amount of items in the result set.
	Items int

	// Page stores the current page number.
	Page int

	// ItemsPerPage stores the amount of items to display per page.
	ItemsPerPage int

	// Pages stores the total amount of pages in the result set.
	Pages int

	// QueryParams stores all query parameters from the request for link generation.
	QueryParams url.Values
}

// NewPager creates a new Pager.
func NewPager(ctx echo.Context, itemsPerPage int) Pager {
	queryParams := ctx.QueryParams()
	p := Pager{
		ItemsPerPage: itemsPerPage,
		Pages:        1,
		Page:         1,
		QueryParams:  make(url.Values), // Initialize the map
	}

	// Deep copy queryParams to avoid modifying the original request's map
	for k, v := range queryParams {
		// http.Request.URL.Query() returns a map[string][]string
		// but echo.Context.QueryParams() returns url.Values which is map[string][]string
		// so direct assignment is fine.
		p.QueryParams[k] = v
	}

	if pageStr := queryParams.Get(QueryKey); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil {
			if pageInt > 0 {
				p.Page = pageInt
			}
		}
	}

	return p
}

// SetItems sets the amount of items in total for the pager and calculate the amount
// of total pages based off on the item per page.
// This should be used rather than setting either items or pages directly.
func (p *Pager) SetItems(items int) {
	p.Items = items

	if items > 0 {
		p.Pages = int(math.Ceil(float64(items) / float64(p.ItemsPerPage)))
	} else {
		p.Pages = 1
	}

	if p.Page > p.Pages {
		p.Page = p.Pages
	}
}

// IsBeginning determines if the pager is at the beginning of the pages
func (p *Pager) IsBeginning() bool {
	return p.Page == 1
}

// IsEnd determines if the pager is at the end of the pages
func (p *Pager) IsEnd() bool {
	return p.Page >= p.Pages
}

// GetOffset determines the offset of the results in order to get the items for
// the current page
func (p *Pager) GetOffset() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.ItemsPerPage
}

// PageURL generates a URL query string for a given page number,
// preserving existing query parameters.
func (p *Pager) PageURL(pageNumber int) string {
	// Create a copy of the existing query parameters to avoid modifying the original
	newParams := make(url.Values)
	for k, v := range p.QueryParams {
		newParams[k] = v
	}

	// Set the desired page number
	if pageNumber <= 1 {
		// If page is 1 or less, remove the page query param for a cleaner URL,
		// assuming page 1 is the default if the param is absent.
		newParams.Del(QueryKey)
	} else {
		newParams.Set(QueryKey, strconv.Itoa(pageNumber))
	}

	if len(newParams) == 0 {
		return ""
	}
	return "?" + newParams.Encode()
}
