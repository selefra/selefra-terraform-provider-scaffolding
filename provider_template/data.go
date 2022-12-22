package provider_template

import _ "embed"

// ------------------------------------------------- --------------------------------------------------------------------

//go:embed schema_custom/custom_schema.tpl
var TableCustomSchemaTemplate string

//go:embed schema_custom/table_list_func.tpl
var TableListFuncTemplate string

//go:embed schema_gen/table_schema_auto_gen.go.tpl
var TableSchemaAutoGenTemplate string

//go:embed schema_gen/table_test.tpl
var TableTableGoTemplate string

// ------------------------------------------------- --------------------------------------------------------------------

//go:embed client/client.go.tpl
var ClientTemplate string

//go:embed schema_manager/schema_manager.go.tpl
var SchemaManagerGoTemplate string

// ------------------------------------------------- --------------------------------------------------------------------

//go:embed provider/provider.go.tpl
var ProviderGoTemplate string

//go:embed provider/provider_test.go.tpl
var ProviderTestGoTemplate string

//go:embed provider/tables.go.tpl
var ProviderTablesGoTemplate string

//go:embed test_provider/test_provider.tpl
var TestProviderGoTemplate string

// ------------------------------------------------- --------------------------------------------------------------------

//go:embed main.go.tpl
var MainTemplate string

// ------------------------------------------------- --------------------------------------------------------------------

//go:embed go.mod.tpl
var GoModTemplate string

// ------------------------------------------------- --------------------------------------------------------------------
