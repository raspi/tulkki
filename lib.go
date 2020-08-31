package tulkki

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"html/template"
	"io"
)

const (
	translationFunctionName = `T`       // Name for a function that can be used to translate inside templates
	baseDefineName          = `base`    // {{define X}} for base template
	pageDefineName          = `content` // {{define X}} for page templates
)

// page is turned into translated html template
type page struct {
	contents string          // Raw HTML as string
	trcat    catalog.Catalog // Translation catalog for this page
}

type pageCache map[language.Tag]map[string]*template.Template

// Template contains HTML used as a base and all the pages that are using that base
type Template struct {
	baseTemplateFuncs    template.FuncMap // Global functions for baseTemplateContents template
	baseTemplateContents string           // Base template as string
	pages                map[string]page  // Pages which uses baseTemplateContents template
	pagesCached          pageCache        // template cache for faster processing
	addPageDefine        bool             // add {{define ...}} to pages automatically?
	pageDefineName       string           // name for define in pages
	baseDefineName       string           // name for define in base
}

// New creates translatable HTML pages
func New(baseContents string, funcs template.FuncMap) Template {
	t := Template{
		pages:             make(map[string]page),
		pagesCached:       make(pageCache),
		baseTemplateFuncs: funcs, // functions available to all pages
		addPageDefine:     true,
		baseDefineName:    baseDefineName,
		pageDefineName:    pageDefineName,
	}

	t.baseTemplateContents = `{{define "` + t.baseDefineName + `"}}` + baseContents + `{{end}}`

	return t
}

// AddPage adds a page which uses baseTemplateContents template as a source.
// Translations are per-page for collision reasons.
// If you need access to translation tokens which are used globally,
// simply inject them to the catalog before page specific tokens.
func (t *Template) AddPage(name string, contents string, translations catalog.Catalog) {
	langlist := translations.Languages()

	if len(langlist) == 0 {
		panic(`no languages found`)
	}

	for _, l := range langlist {
		// Init cache for pages
		t.pagesCached[l] = make(map[string]*template.Template)
	}

	if t.addPageDefine {
		contents = `{{define "` + t.pageDefineName + `"}}` + contents + `{{end}}`
	}

	t.pages[name] = page{
		contents: contents,
		trcat:    translations,
	}
}

func (t *Template) getPageTemplate(templatename string) (p page, err error) {
	p, ok := t.pages[templatename]

	if !ok {
		return p, fmt.Errorf(`couldn't load template named %q`, templatename)
	}

	return p, nil
}

// getTemplate gets page's *template.Template in given language
func (t *Template) getTemplate(templatename string, language language.Tag) (tpl *template.Template, err error) {
	// Fetch template from cache
	tpl, ok := t.pagesCached[language][templatename]

	if ok {
		return tpl, nil
	}

	// Template was not cached, generate
	tpl = template.New(``)

	// Load common template functions
	_ = tpl.Funcs(t.baseTemplateFuncs)

	pageTemplate, err := t.getPageTemplate(templatename)
	if err != nil {
		return tpl, err
	}

	_ = tpl.Funcs(template.FuncMap{
		// Translate inside template
		translationFunctionName: func(s string, a ...interface{}) template.HTML {
			pr := message.NewPrinter(language, message.Catalog(pageTemplate.trcat))
			return template.HTML(pr.Sprintf(s, a...))
		},
	})

	_, err = tpl.Parse(t.baseTemplateContents)
	if err != nil {
		return tpl, fmt.Errorf(`couldn't parse base template: %w`, err)
	}

	_, err = tpl.Parse(pageTemplate.contents)
	if err != nil {
		return tpl, fmt.Errorf(`couldn't parse template %q: %w`, templatename, err)
	}

	// Save to cache
	t.pagesCached[language][templatename] = tpl

	return tpl, nil
}

// Render outputs generated HTML to a writer
func (t *Template) Render(w io.Writer, templatename string, language language.Tag, data interface{}) (err error) {
	tpl, err := t.getTemplate(templatename, language)
	if err != nil {
		return fmt.Errorf(`couldn't get template %q: %w`, templatename, err)
	}

	err = tpl.ExecuteTemplate(w, baseDefineName, data)
	if err != nil {
		return fmt.Errorf(`couldn't execute template %q: %w`, templatename, err)
	}

	return nil
}

// Translate a token with given language
func (t *Template) Translate(templatename string, key string, language language.Tag, data ...interface{}) string {
	pageTemplate, err := t.getPageTemplate(templatename)
	if err != nil {
		panic(err)
	}

	tr := message.NewPrinter(language, message.Catalog(pageTemplate.trcat))
	return tr.Sprintf(key, data...)
}
