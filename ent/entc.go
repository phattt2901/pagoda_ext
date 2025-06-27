//go:build ignore
// +build ignore

package main

import (
	"log"
	"regexp"
	"strings"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"text/template" // Required for template.FuncMap
	"github.com/mikestefanello/pagoda/ent/admin"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake  = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func main() {
	// Custom template functions
	funcs := template.FuncMap{
		"camelToSnake": camelToSnake,
	}

	err := entc.Generate("./schema",
		&gen.Config{
			Funcs: funcs, // Add custom functions here
		},
		entc.Extensions(&admin.Extension{}),
	)
	if err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
