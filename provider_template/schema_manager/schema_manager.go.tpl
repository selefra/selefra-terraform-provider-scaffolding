package schema_manager

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/table_schema_generator"
)

var tableSchemaGeneratorMap = make(map[string]table_schema_generator.TableSchemaGenerator)

func RegisterTable(tableName string, tableSchemaGenerator table_schema_generator.TableSchemaGenerator) {
	tableSchemaGeneratorMap[tableName] = tableSchemaGenerator
}

func FindTables(tableNames []string) ([]*schema.Table, error) {
	tableSlice := make([]*schema.Table, 0)
	for _, tableName := range tableNames {
		tableSchemaGenerator, exists := tableSchemaGeneratorMap[tableName]
		if !exists {
			return nil, fmt.Errorf("table %s not found", tableName)
		}
		tableSlice = append(tableSlice, table_schema_generator.GenTableSchema(tableSchemaGenerator))
	}
	return tableSlice, nil
}
