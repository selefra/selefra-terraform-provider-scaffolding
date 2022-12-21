package tables

import (
	"github.com/selefra/selefra-provider-sdk/table_schema_generator"
	"github.com/selefra/selefra-provider-sdk/test_helper"
	"{{.GoModuleUrl}}/test_provider"
	"testing"
)

func Test_{{.ResourceName}}(t *testing.T) {
	testProvider := test_provider.GetProvider()
	testProvider.TableList = append(testProvider.TableList, table_schema_generator.GenTableSchema(&{{.TableGeneratorName}}{}))
	config := "test : test"
	test_helper.RunProviderPullTables(testProvider, config, "./", "*")
}
