package tulkki

import (
	"bytes"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"html/template"
	"testing"
)

func getTestTranslations() *catalog.Builder {
	translations := catalog.NewBuilder()

	translations.SetString(language.English, `t`, `test`)
	translations.SetString(language.Finnish, `t`, `testi`)

	return translations
}

func TestBaseTemplateTranslationOnly(t *testing.T) {
	translations := getTestTranslations()

	tpl := New(`<html>{{- T "t" -}}</html>`, template.FuncMap{})
	tpl.AddPage(`test`, ``, translations)

	var tmp bytes.Buffer
	for _, l := range translations.Languages() {
		tmp.Reset()

		err := tpl.Render(&tmp, `test`, l, nil)
		if err != nil {
			t.Fail()
		}

		printer := message.NewPrinter(l, message.Catalog(translations))
		expected := `<html>` + printer.Sprintf(`t`) + `</html>`

		if expected != tmp.String() {
			t.Fail()
		}
	}
}

func TestPageTranslationFunction(t *testing.T) {
	translations := getTestTranslations()

	tpl := New(`<html>{{- template "content" .C -}}</html>`, template.FuncMap{})
	tpl.AddPage(`test`, `{{- T "t" -}}`, translations)

	var tmp bytes.Buffer
	for _, l := range translations.Languages() {
		tmp.Reset()

		err := tpl.Render(&tmp, `test`, l, struct {
			C interface{}
		}{})
		if err != nil {
			t.Fail()
		}

		printer := message.NewPrinter(l, message.Catalog(translations))
		expected := `<html>` + printer.Sprintf(`t`) + `</html>`

		if expected != tmp.String() {
			t.Fail()
		}
	}
}

func TestPageTranslationVariable(t *testing.T) {
	translations := getTestTranslations()

	tpl := New(`<html>{{- template "content" .C -}}</html>`, template.FuncMap{})
	tpl.AddPage(`test`, `{{- .T -}}`, translations)

	var tmp bytes.Buffer
	for _, l := range translations.Languages() {
		tmp.Reset()

		err := tpl.Render(&tmp, `test`, l, struct {
			C interface{}
		}{
			C: struct {
				T string
			}{
				T: tpl.Translate(`test`, `t`, l, nil),
			},
		})
		if err != nil {
			t.Fail()
		}

		printer := message.NewPrinter(l, message.Catalog(translations))
		expected := `<html>` + printer.Sprintf(`t`) + `</html>`

		if expected != tmp.String() {
			t.Fail()
		}
	}
}
