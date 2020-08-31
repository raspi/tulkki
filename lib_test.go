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
	err := tpl.AddPage(`test`, ``, translations)
	if err != nil {
		t.Fail()
	}

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
	err := tpl.AddPage(`test`, `{{- T "t" -}}`, translations)
	if err != nil {
		t.Fail()
	}

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
	err := tpl.AddPage(`test`, `{{- .T -}}`, translations)
	if err != nil {
		t.Fail()
	}

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

func TestCache(t *testing.T) {
	translations := getTestTranslations()

	tpl := New(`<html>{{- T "t" -}}</html>`, template.FuncMap{})
	err := tpl.AddPage(`test`, ``, translations)
	if err != nil {
		t.Fail()
	}

	var tmp bytes.Buffer

	for i := 0; i < 10; i++ {
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
}

func TestPageTokenTranslation(t *testing.T) {
	translations := getTestTranslations()

	tpl := New(``, template.FuncMap{})
	err := tpl.AddPage(`test`, ``, translations)
	if err != nil {
		t.Fail()
	}

	for _, l := range translations.Languages() {
		actual := tpl.Translate(`test`, `t`, l)

		printer := message.NewPrinter(l, message.Catalog(translations))
		expected := printer.Sprintf(`t`)

		if expected != actual {
			t.Fail()
		}
	}
}
