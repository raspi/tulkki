package main

import (
	"bytes"
	"fmt"
	"github.com/raspi/tulkki"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"html/template"
	"time"
)

// Base template for all pages
var baseHTML = `
<html>
<head>
<title>{{.Title}}</title>
</head>
<body>
<main>
{{- template "content" .C -}}
</main>
</body>
</html>
`

// Example page
var testPageHTML = `
<p>
  <!-- Translated within template itself: -->
  {{T "test"}}
  {{T "test_formatted" "Template" "right now"}}
  <!-- Pre-translated to a template variable: -->
  {{.TranslatedToken}}
</p>
`

// example base struct which all pages will use
type exampleBase struct {
	Title string
	// C (short for content) is used in the base HTML template as a per-page variables
	// It needs to be abstract interface because different pages have different variables
	C interface{}
}

// example page
type examplePage struct {
	TranslatedToken string
}

func main() {
	useLanguage := language.English
	fallbackLanguage := language.English

	pagetranslations := catalog.NewBuilder(
		catalog.Fallback(fallbackLanguage),
	)

	// Translations are per-page for collision reasons
	pagetranslations.SetString(language.English, `test`, `this is a test string`)
	pagetranslations.SetString(language.English, `test_formatted`, `%s is testing things on %s`)

	// Global template functions accessible to all pages
	// Add CSRF etc generators here
	funcs := template.FuncMap{}

	tpl := tulkki.New(baseHTML, funcs)
	err := tpl.AddPage(`testpage`, testPageHTML, pagetranslations)
	if err != nil {
		panic(err)
	}

	// Generate a page template
	page := exampleBase{
		Title: "Hello, world!",
		C: examplePage{
			TranslatedToken: tpl.Translate(`testpage`, `test_formatted`, useLanguage, `Example`, time.Now().Truncate(time.Second)),
		},
	}

	// Render the HTML
	var tmp bytes.Buffer
	err = tpl.Render(&tmp, `testpage`, useLanguage, page)
	if err != nil {
		panic(err)
	}

	fmt.Print(tmp.String())

	fmt.Println(`example: ` + tpl.Translate(`testpage`, `test_formatted`, useLanguage, `Direct translation`, template.HTML(`<a href="#">here</a>`)))
}
