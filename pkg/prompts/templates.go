// Package prompts provides prompt template management for LLM interactions
package prompts

import (
	"fmt"
	"strings"
)

// Template represents a reusable prompt template with variable substitution
type Template struct {
	Name     string
	Template string
	Vars     []string // required variable names
}

// NewTemplate creates a new prompt template
func NewTemplate(name, tmpl string, vars []string) *Template {
	return &Template{
		Name:     name,
		Template: tmpl,
		Vars:     vars,
	}
}

// Render substitutes variables into the template
// Variables are referenced as {var_name} in the template string
func (t *Template) Render(vars map[string]string) (string, error) {
	// Check required vars are present
	for _, v := range t.Vars {
		if _, ok := vars[v]; !ok {
			return "", fmt.Errorf("template %s: missing required variable '%s'", t.Name, v)
		}
	}

	result := t.Template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{"+k+"}", v)
	}

	return result, nil
}

// MustRender renders the template, panicking on error (for static templates)
func (t *Template) MustRender(vars map[string]string) string {
	result, err := t.Render(vars)
	if err != nil {
		panic(err)
	}
	return result
}

// TemplateRegistry holds a collection of named templates
type TemplateRegistry struct {
	templates map[string]*Template
}

// NewRegistry creates an empty template registry
func NewRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*Template),
	}
}

// Register adds a template to the registry
func (r *TemplateRegistry) Register(t *Template) {
	r.templates[t.Name] = t
}

// Get retrieves a template by name
func (r *TemplateRegistry) Get(name string) (*Template, bool) {
	t, ok := r.templates[name]
	return t, ok
}

// MustGet retrieves a template or panics if not found
func (r *TemplateRegistry) MustGet(name string) *Template {
	t, ok := r.templates[name]
	if !ok {
		panic(fmt.Sprintf("template '%s' not found in registry", name))
	}
	return t
}

// DefaultRegistry is the global prompt template registry
var DefaultRegistry = NewRegistry()

func init() {
	// Register all built-in templates
	for _, t := range extractionTemplates {
		DefaultRegistry.Register(t)
	}
	for _, t := range reasoningTemplates {
		DefaultRegistry.Register(t)
	}
}
