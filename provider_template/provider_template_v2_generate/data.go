package provider_template_v2_generate

import _ "embed"

//go:embed selefra_schema.go.tpl
var SelefraSchemaTemplate string

//go:embed selefra_provider.go.tpl
var SelefraProviderTemplate string

//go:embed selefra_provider_test.go.tpl
var SelefraProviderTestTemplate string

//go:embed main.go.tpl
var MainTemplate string

//go:embed go.mod.tpl
var GoModTemplate string
