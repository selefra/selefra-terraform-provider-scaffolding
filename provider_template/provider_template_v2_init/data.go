package provider_template_v2_init

import _ "embed"

//go:embed provider.go.tpl
var ProviderTemplate string

//go:embed resources.go.tpl
var ResourceTemplate string
