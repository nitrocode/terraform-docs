/*
Copyright 2021 The terraform-docs Authors.

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/

package format

import (
	gotemplate "text/template"

	"github.com/terraform-docs/terraform-docs/internal/print"
	"github.com/terraform-docs/terraform-docs/internal/template"
	"github.com/terraform-docs/terraform-docs/internal/terraform"
)

const (
	tableHeaderTpl = `
	{{- if .Settings.ShowHeader -}}
		{{- with .Module.Header -}}
			{{ sanitizeHeader . }}
			{{ printf "\n" }}
		{{- end -}}
	{{ end -}}
	`
	tableResourcesTpl = `
	{{- if .Settings.ShowResources -}}
		{{ indent 0 "#" }} Resources
		{{ if not .Module.Resources }}
			No resources.
		{{ else }}
			| Name |
			|------|
			{{- range .Module.Resources }}
			<a name="resources_{{ .FullType }}"></a>
			{{ if eq (len .URL) 0 }}
				| {{ .FullType }}
			{{- else -}}
				| [{{ .FullType }}]({{ .URL }}) |
			{{- end }}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableRequirementsTpl = `
	{{- if .Settings.ShowRequirements -}}
		{{ indent 0 "#" }} Requirements
		{{ if not .Module.Requirements }}
			No requirements.
		{{ else }}
			| Name | Version |
			|------|---------|
			{{- range .Module.Requirements }}
				| <a name="requirements_{{ .Name }}"></a> {{ name .Name }} | {{ tostring .Version | default "n/a" }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableProvidersTpl = `
	{{- if .Settings.ShowProviders -}}
		{{ indent 0 "#" }} Providers
		{{ if not .Module.Providers }}
			No provider.
		{{ else }}
			| Name | Version |
			|------|---------|
			{{- range .Module.Providers }}
				| <a name="providers_{{ .FullName }}"></a> {{ name .FullName }} | {{ tostring .Version | default "n/a" }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableInputsTpl = `
	{{- if .Settings.ShowInputs -}}
		{{ indent 0 "#" }} Inputs
		{{ if not .Module.Inputs }}
			No input.
		{{ else }}
			| Name | Description | Type | Default |{{ if .Settings.ShowRequired }} Required |{{ end }}
			|------|-------------|------|---------|{{ if .Settings.ShowRequired }}:--------:|{{ end }}
			{{- range .Module.Inputs }}
				| <a name="inputs_{{ .Name }}"></a> {{ name .Name }} | {{ tostring .Description | sanitizeTbl }} | {{ tostring .Type | type | sanitizeTbl }} | {{ value .GetValue | sanitizeTbl }} |
				{{- if $.Settings.ShowRequired -}}
					{{ printf " " }}{{ ternary .Required "yes" "no" }} |
				{{- end -}}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableOutputsTpl = `
	{{- if .Settings.ShowOutputs -}}
		{{ indent 0 "#" }} Outputs
		{{ if not .Module.Outputs }}
			No output.
		{{ else }}
			| Name | Description |{{ if .Settings.OutputValues }} Value |{{ if $.Settings.ShowSensitivity }} Sensitive |{{ end }}{{ end }}
			|------|-------------|{{ if .Settings.OutputValues }}-------|{{ if $.Settings.ShowSensitivity }}:---------:|{{ end }}{{ end }}
			{{- range .Module.Outputs }}
				| <a name="outputs_{{ .Name }}"></a> {{ name .Name }} | {{ tostring .Description | sanitizeTbl }} |
				{{- if $.Settings.OutputValues -}}
					{{- $sensitive := ternary .Sensitive "<sensitive>" .GetValue -}}
					{{ printf " " }}{{ value $sensitive | sanitizeTbl }} |
					{{- if $.Settings.ShowSensitivity -}}
						{{ printf " " }}{{ ternary .Sensitive "yes" "no" }} |
					{{- end -}}
				{{- end -}}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableModulecallsTpl = `
	{{- if .Settings.ShowModuleCalls -}}
		{{ indent 0 "#" }} Modules
		{{ if not .Module.ModuleCalls }}
			No Modules.
		{{ else }}
			| Name | Source | Version |
			|------|--------|---------|
			{{- range .Module.ModuleCalls }}
				| <a name="modules_{{ .Name }}"></a> {{ .Name }} | {{ .Source }} | {{ .Version }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	tableTpl = `
	{{- template "header" . -}}
	{{- template "requirements" . -}}
	{{- template "providers" . -}}
	{{- template "modulecalls" . -}}
	{{- template "resources" . -}}
	{{- template "inputs" . -}}
	{{- template "outputs" . -}}
	`
)

// MarkdownTable represents Markdown Table format.
type MarkdownTable struct {
	template *template.Template
}

// NewMarkdownTable returns new instance of Table.
func NewMarkdownTable(settings *print.Settings) print.Engine {
	tt := template.New(settings, &template.Item{
		Name: "table",
		Text: tableTpl,
	}, &template.Item{
		Name: "header",
		Text: tableHeaderTpl,
	}, &template.Item{
		Name: "requirements",
		Text: tableRequirementsTpl,
	}, &template.Item{
		Name: "providers",
		Text: tableProvidersTpl,
	}, &template.Item{
		Name: "resources",
		Text: tableResourcesTpl,
	}, &template.Item{
		Name: "inputs",
		Text: tableInputsTpl,
	}, &template.Item{
		Name: "outputs",
		Text: tableOutputsTpl,
	}, &template.Item{
		Name: "modulecalls",
		Text: tableModulecallsTpl,
	})
	tt.CustomFunc(gotemplate.FuncMap{
		"type": func(t string) string {
			inputType, _ := printFencedCodeBlock(t, "")
			return inputType
		},
		"value": func(v string) string {
			var result = "n/a"
			if v != "" {
				result, _ = printFencedCodeBlock(v, "")
			}
			return result
		},
	})
	return &MarkdownTable{
		template: tt,
	}
}

// Print a Terraform module as Markdown tables.
func (t *MarkdownTable) Print(module *terraform.Module, settings *print.Settings) (string, error) {
	rendered, err := t.template.Render(module)
	if err != nil {
		return "", err
	}
	return sanitize(rendered), nil
}

func init() {
	register(map[string]initializerFn{
		"markdown":       NewMarkdownTable,
		"markdown table": NewMarkdownTable,
		"markdown tbl":   NewMarkdownTable,
		"md":             NewMarkdownTable,
		"md table":       NewMarkdownTable,
		"md tbl":         NewMarkdownTable,
	})
}
