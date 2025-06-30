package admin

import (
	"embed"
	"regexp" // Added for camelToSnake
	"strings"
	"text/template"
	"unicode"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
)

var (
	//go:embed templates
	templateDir embed.FS
)

// Extension is the Ent extension that generates code to support the entity admin panel.
type Extension struct {
	entc.DefaultExtension
}

func (*Extension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(
			gen.NewTemplate("admin").
				Funcs(template.FuncMap{
					"fieldName":      fieldName,
					"fieldLabel":     FieldLabel,
					"fieldIsPointer": fieldIsPointer,
					"camelToSnake":   camelToSnake, // Added camelToSnake
				}).
				ParseFS(templateDir, "templates/*tmpl"),
		),
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

// camelToSnake converts a CamelCase string to snake_case.
func camelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake  = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// fieldName provides a struct field name from an entity field name (ie, user_id -> UserID).
func fieldName(name string) string {
	if len(name) == 0 {
		return name
	}

	parts := strings.Split(name, "_")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "id" {
			parts[i] = "ID"
		} else {
			parts[i] = upperFirst(parts[i])
		}
	}

	return strings.Join(parts, "")
}

// FieldLabel provides a label for an entity field name (ie, user_id -> User ID).
func FieldLabel(name string) string {
	if len(name) == 0 {
		return name
	}

	parts := strings.Split(name, "_")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "id" {
			parts[i] = "ID"
		}
		if i == 0 {
			parts[i] = upperFirst(parts[i])
		}
	}

	return strings.Join(parts, " ")
}

// fieldIsPointer determines if a given entity field should be a pointer on the struct.
func fieldIsPointer(f *gen.Field) bool {
	switch {
	case f.Type.Type == field.TypeBool:
		return false
	case f.Optional,
		f.Default,
		f.Sensitive(),
		f.Nillable:
		return true
	}
	return false
}

// upperFirst uppercases the first character of a given string.
func upperFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	out := []rune(s)
	out[0] = unicode.ToUpper(out[0])
	return string(out)
}
