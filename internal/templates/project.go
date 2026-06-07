package templates

import _ "embed"

//go:embed gox.yaml.tmpl
var GoxYamlTemplate string

//go:embed app_main.go.tmpl
var AppMainTemplate string

//go:embed lib.go.tmpl
var LibTemplate string
