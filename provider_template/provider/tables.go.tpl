package provider

import (
		"github.com/selefra/selefra-provider-sdk/provider/schema"
    	"github.com/selefra/selefra-provider-sdk/table_schema_generator"
    	"{{.ModuleName}}/schema_gen"
)

func GenTables() []*schema.Table {
	return []*schema.Table{ {{range $index, $value := .TableGeneratorNameSlice }}
	    table_schema_generator.GenTableSchema(&schema_gen.{{$value}}{}), {{end}}
	}
}