package provider_template_v2_generate

import _ "embed"

//go:embed schema.go.tpl
var SchemaTemplate string

//go:embed provider.go.tpl
var ProviderTemplate string

//go:embed selefra_provider_main/main.go.tpl
var MainTemplate string

//go:embed go.mod.tpl
var GoModTemplate string
