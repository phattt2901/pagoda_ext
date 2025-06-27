package pages

import (
	"fmt"
	"net/url"

	"entgo.io/ent/entc/load"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent/admin"
	"github.com/mikestefanello/pagoda/pkg/pager" // Added pager import
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "github.com/mikestefanello/pagoda/pkg/ui/components"
	"github.com/mikestefanello/pagoda/pkg/ui/forms"
	"github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func AdminEntityDelete(ctx echo.Context, entityTypeName string) error {
	r := ui.NewRequest(ctx)
	r.Title = fmt.Sprintf("Delete %s", entityTypeName)

	return r.Render(
		layouts.Primary,
		forms.AdminEntityDelete(r, entityTypeName),
	)
}

func AdminEntityInput(ctx echo.Context, schema *load.Schema, values url.Values) error {
	r := ui.NewRequest(ctx)
	if values == nil {
		r.Title = fmt.Sprintf("Add %s", schema.Name)
	} else {
		r.Title = fmt.Sprintf("Edit %s", schema.Name)
	}

	return r.Render(
		layouts.Primary,
		forms.AdminEntity(r, schema, values),
	)
}

func AdminEntityList(
	ctx echo.Context,
	entityTypeName string,
	entityList *admin.EntityList,
) error {
	r := ui.NewRequest(ctx)
	r.Title = entityTypeName

	genHeader := func() Node {
		g := make(Group, 0, len(entityList.Columns)+2)
		g = append(g, Th(Text("ID")))
		for _, h := range entityList.Columns {
			g = append(g, Th(Text(h)))
		}
		g = append(g, Th())
		return g
	}

	genRow := func(row admin.EntityValues) Node {
		g := make(Group, 0, len(row.Values)+3)
		g = append(g, Th(Text(fmt.Sprint(row.ID))))
		for _, h := range row.Values {
			g = append(g, Td(Text(h)))
		}
		g = append(g,
			Td(
				ButtonLink(
					ColorInfo,
					r.Path(routenames.AdminEntityEdit(entityTypeName), row.ID),
					"Edit",
				),
				Span(Class("mr-2")),
				ButtonLink(
					ColorError,
					r.Path(routenames.AdminEntityDelete(entityTypeName), row.ID),
					"Delete",
				),
			),
		)
		return g
	}

	genRows := func() Node {
		g := make(Group, 0, len(entityList.Entities))
		for _, row := range entityList.Entities {
			g = append(g, Tr(genRow(row)))
		}
		return g
	}

	return r.Render(layouts.Primary, Group{
		Div(
			Class("form-control mb-2"),
			ButtonLink(
				ColorAccent,
				r.Path(routenames.AdminEntityAdd(entityTypeName)),
				fmt.Sprintf("Add %s", entityTypeName),
			),
		),
		Table(
			Class("table table-zebra mb-2"),
			THead(
				Tr(genHeader()),
			),
			TBody(genRows()),
		),
		renderAdminPager(r, entityList), // Updated Pager call
	})
}

// renderAdminPager helper function to encapsulate pager logic
func renderAdminPager(r *ui.Request, entityList *admin.EntityList) Node {
	// This is a temporary workaround. Ideally, the handler creates the pager
	// and passes necessary info (TotalItems, ItemsPerPage) via entityList or another way.
	// For now, we'll use a default ItemsPerPage and estimate TotalItems if not available.
	const itemsPerPage = 25 // Default, should come from config or handler

	p := pager.NewPager(r.Context, itemsPerPage)

	// entityList.TotalItems is not currently set by the generated handler.
	// We only have HasNextPage. This makes accurate total page count difficult here.
	// The Pager component itself primarily cares about prev/next links.
	// For a full display of "Page X of Y", TotalItems is needed.
	// Let's assume for now we only enable/disable based on HasNextPage and current page.
	// p.SetItems(entityList.TotalItems) // This would be ideal

	currentPage := entityList.Page
	hasNext := entityList.HasNextPage

	prevPageURL := ""
	if currentPage > 1 {
		prevPageURL = p.PageURL(currentPage - 1)
	}

	nextPageURL := ""
	if hasNext {
		nextPageURL = p.PageURL(currentPage + 1)
	}

	// The hxTarget for admin panel lists would typically be the main content area
	// holding the table and pager itself, if HTMX is used for pagination.
	// For now, an empty string means normal links.
	hxTarget := "" // Example: "#admin-entity-list-container"

	return Pager(
		currentPage,
		prevPageURL,
		nextPageURL,
		hasNext,
		hxTarget,
	)
}
