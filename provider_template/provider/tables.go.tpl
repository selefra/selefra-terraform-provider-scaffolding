package provider

import (
		"github.com/selefra/selefra-provider-sdk/provider/schema"
    	"github.com/selefra/selefra-provider-sdk/table_schema_generator"
    	"{{.GoModuleName}}/tables"
    	{{range $index, $value := .ImportSlice}}
    {{$value}} {{end}}
)

func GenTables() []*schema.Table {
	return []*schema.Table{ {{range $index, $value := .TableGeneratorNameSlice }}
	    table_schema_generator.GenTableSchema(&tables.{{$value}}{}), {{end}}
	}
}