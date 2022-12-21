package provider_template

import _ "embed"

// ResourceTableSchemaTemplate Resource对应的表的模板
//
//go:embed tables/table_schema_auto_gen.go.tpl
var ResourceTableSchemaTemplate string

// ResourceTableListTemplate 表所对应的list函数的模板
//
//go:embed tables/table_list_func.tpl
var ResourceTableListTemplate string

//go:embed client/client.go.tpl
var ClientTemplate string

//go:embed provider/provider.go.tpl
var ProviderTemplate string

//go:embed provider/tables.go.tpl
var ProviderTablesTemplate string

//go:embed go.mod.tpl
var GoModTemplate string

//go:embed test_provider/test_provider.tpl
var TestProvider string

//go:embed tables/table_test.tpl
var TableTest string
