package components

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/pager"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func MenuLink(r *ui.Request, icon Node, title, routeName string, routeParams ...any) Node {
	href := r.Path(routeName, routeParams...)

	return Li(
		Class("ml-2"),
		A(
			Href(href),
			icon,
			Text(title),
			Classes{
				"menu-active": href == r.CurrentPath,
				"p-2":         true,
			},
		),
	)
}

// Pager generates pagination controls.
// currentPage is the current page number.
// prevPageURL is the URL for the previous page; empty if no previous page.
// nextPageURL is the URL for the next page; empty if no next page.
// hasNext is true if there is a next page (used to enable/disable next button more explicitly than just checking nextPageURL).
// hxTarget is an optional htmx target for AJAX pagination.
func Pager(currentPage int, prevPageURL string, nextPageURL string, hasNext bool, hxTarget string) Node {
	return Div(
		Class("join"),
		A(
			Class("join-item btn"),
			Text("«"), // Previous
			If(currentPage <= 1 || prevPageURL == "", Disabled()),
			Href(prevPageURL), // Use pre-generated URL
			Iff(len(hxTarget) > 0 && prevPageURL != "", func() Node {
				return Group{
					Attr("hx-get", prevPageURL),
					Attr("hx-swap", "outerHTML"),
					Attr("hx-target", hxTarget),
				}
			}),
		),
		Button(
			Class("join-item btn"),
			Textf("Page %d", currentPage),
		),
		A(
			Class("join-item btn"),
			Text("»"), // Next
			If(!hasNext || nextPageURL == "", Disabled()),
			Href(nextPageURL), // Use pre-generated URL
			Iff(len(hxTarget) > 0 && nextPageURL != "" && hasNext, func() Node {
				return Group{
					Attr("hx-get", nextPageURL),
					Attr("hx-swap", "outerHTML"),
					Attr("hx-target", hxTarget),
				}
			}),
		),
	)
}
