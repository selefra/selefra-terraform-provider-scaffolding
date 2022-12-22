package generate_selefra_terraform_provider

import (
	"bytes"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"github.com/yezihack/colorlog"
	"os"
	"strings"
	"text/template"
)

// TableGenerator Generate resources related to terraformResourceSchemaInfo
type TableGenerator struct {

	// The configuration information is generated
	config *Config

	// the provider to which the current terraformResourceSchemaInfo belongs
	provider *ProviderGenerator

	// The name of the terraformResourceSchemaInfo currently being processed
	tableName                   string
	terraformResourceSchemaInfo shim.Resource

	// Parameters when the table is generated from the terraformResourceSchemaInfo
	params *TableSchemaAutoGenRenderParams
}

// NewResourceGenerator Create a terraformResourceSchemaInfo to table generator
func NewResourceGenerator(config *Config, provider *ProviderGenerator, tableName string, terraformResourceSchemaInfo shim.Resource) *TableGenerator {
	return &TableGenerator{
		config:                      config,
		provider:                    provider,
		tableName:                   tableName,
		terraformResourceSchemaInfo: terraformResourceSchemaInfo,
		params:                      NewTableSchemaAutoGenRenderParams(),
	}
}

func (x *TableGenerator) Run() error {

	// Generate table names and the like
	x.params.TableName = x.tableName
	x.params.TableSchemaGeneratorName = snakeToCamel(x.tableName) + "Generator"
	x.params.ModuleName = x.config.Selefra.ModuleName

	// Generate the columns of the table
	if isHaveIdColumn, err := x.GenSelefraColumnsInformationFromTerraform(); err != nil || !isHaveIdColumn {
		return err
	}

	// rendering template
	if err := x.RenderAndSave(); err != nil {
		return err
	}

	return nil
}

// Underline the hump
func snakeToCamel(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

func (x *TableGenerator) RenderAndSave() error {

	// step 1. generate auto schema
	buffer := bytes.Buffer{}
	t, err := template.New("table-schema-auto-gen").Parse(string(provider_template.TableSchemaAutoGenTemplate))
	if err != nil {
		return err
	}
	if err = t.ExecuteTemplate(&buffer, "table-schema-auto-gen", x.params); err != nil {
		return err
	}
	tableSchemaAutoGenOutputDirectory := x.config.Output.Directory + "/schema_gen/"
	_ = os.MkdirAll(tableSchemaAutoGenOutputDirectory, os.ModePerm)
	tableSchemaAutoGenOutputFile := tableSchemaAutoGenOutputDirectory + x.tableName + "__schema.go"
	if err := os.WriteFile(tableSchemaAutoGenOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// step 2. And then the list function
	tableCustomSchemaOutputDirectory := x.config.Output.Directory + "/schema_custom/"
	_ = os.MkdirAll(tableCustomSchemaOutputDirectory, os.ModePerm)
	tableCustomSchemaListOutputFile := tableCustomSchemaOutputDirectory + x.tableName + "__list.go"
	t, err = template.New("list.go").Parse(string(provider_template.TableListFuncTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "list.go", &ResourceListParams{
		TerraformProviderShortName: x.config.Terraform.TerraformProvider.ParseProviderShortName(),
		TableName:                  x.tableName,
	}); err != nil {
		return err
	}
	if exists, err := PathExists(tableCustomSchemaListOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", tableCustomSchemaListOutputFile)
	} else {
		if err := os.WriteFile(tableCustomSchemaListOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
			return err
		}
	}

	// step 3. generate custom schema
	t, err = template.New("table-custom-schema").Parse(string(provider_template.TableCustomSchemaTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "table-custom-schema", SchemaCustomRenderParams{
		TableName: x.tableName,
	}); err != nil {
		return err
	}
	tableCustomSchemaOutputFile := tableCustomSchemaOutputDirectory + x.tableName + "__schema.go"
	if exists, err := PathExists(tableCustomSchemaOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", tableCustomSchemaOutputFile)
	} else {
		if err := os.WriteFile(tableCustomSchemaOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
			return err
		}
	}

	// step 4. And table test file
	tableTestOutputFile := tableSchemaAutoGenOutputDirectory + x.tableName + "__test.go"
	t, err = template.New("table-test").Parse(string(provider_template.TableTableGoTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "table-test", &TableTestParams{
		TableName:          x.tableName,
		ModuleName:         x.config.Selefra.ModuleName,
		TableGeneratorName: x.params.TableSchemaGeneratorName,
	}); err != nil {
		return err
	}
	if err := os.WriteFile(tableTestOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// Rendering is successful. Report yourself
	x.provider.TableGeneratorNameSlice = append(x.provider.TableGeneratorNameSlice, x.params.TableSchemaGeneratorName)

	return nil
}

// GenSelefraColumnsInformationFromTerraform Generate the columns corresponding to the table
func (x *TableGenerator) GenSelefraColumnsInformationFromTerraform() (bool, error) {
	haveIdColumn := false
	x.terraformResourceSchemaInfo.Schema().Range(func(columnName string, value shim.Schema) bool {
		column := &ColumnSchema{
			ColumnName:  columnName,
			Description: value.Description(),
		}

		// column's type & column value extractor
		switch value.Type() {
		case shim.TypeBool:
			column.ColumnTypeCodeString = "schema.ColumnTypeBool"
		case shim.TypeInt:
			column.ColumnTypeCodeString = "schema.ColumnTypeBigInt"
		case shim.TypeFloat:
			column.ColumnTypeCodeString = "schema.ColumnTypeFloat"
		case shim.TypeString:
			column.ColumnTypeCodeString = "schema.ColumnTypeString"
		case shim.TypeList, shim.TypeMap, shim.TypeSet:
			// All are converted to JSON
			column.ColumnTypeCodeString = "schema.ColumnTypeJSON"
			column.ExtractorInlineCodeString = "column_value_extractor.TerraformRawDataColumnValueExtractor()"
			x.params.ImportSlice["github.com/selefra/selefra-provider-sdk/terraform/column_value_extractor"] = struct{}{}
		case shim.TypeInvalid:
			column.ColumnTypeCodeString = "schema.ColumnTypeInvalid"
		}

		// primary key
		if columnName == "id" {
			x.params.PrimaryKeys = append(x.params.PrimaryKeys, "id")
			haveIdColumn = true
		}

		x.params.ColumnSchemaSlice = append(x.params.ColumnSchemaSlice, column)

		return true
	})

	if !haveIdColumn {
		return false, nil
	}

	// Add an additional column to store the original response data
	x.params.ColumnSchemaSlice = append(x.params.ColumnSchemaSlice, &ColumnSchema{
		ColumnName:                "selefra_terraform_original_result",
		Description:               "save terraform original result for compatibility",
		ColumnTypeCodeString:      "schema.ColumnTypeJSON",
		ExtractorInlineCodeString: "column_value_extractor.TerraformRawDataColumnValueExtractor()",
	})
	x.params.ImportSlice["github.com/selefra/selefra-provider-sdk/terraform/column_value_extractor"] = struct{}{}

	return true, nil
}

// ------------------------------------------------ ---------------------------------------------------------------------

// TableSchemaAutoGenRenderParams Parameters needed to render the template
type TableSchemaAutoGenRenderParams struct {

	// The name of the table generator
	TableSchemaGeneratorName string

	// Name of table
	TableName string

	// Description of the table
	Description string

	// All the columns in the table
	ColumnSchemaSlice []*ColumnSchema

	// Primary key columns in a table
	PrimaryKeys []string

	// The current table generator depends on which packages need to be imported
	ImportSlice map[string]struct{}

	ModuleName string
}

func NewTableSchemaAutoGenRenderParams() *TableSchemaAutoGenRenderParams {
	return &TableSchemaAutoGenRenderParams{
		ImportSlice: make(map[string]struct{}, 0),
	}
}

// ------------------------------------------------ ---------------------------------------------------------------------

// ColumnSchema Represents structural information about a column in a table
type ColumnSchema struct {

	// Name of column
	ColumnName string

	// Description of the column
	Description string

	// The code snippet corresponding to the column type
	ColumnTypeCodeString string

	// The code snippet corresponding to the column extractor
	ExtractorInlineCodeString string

	// Option at column creation time
	Options schema.ColumnOptions
}

// ResourceListParams Parameter to generate a List terraformResourceSchemaInfo
type ResourceListParams struct {

	// Name of the terraform provider
	TerraformProviderShortName string

	// list Specifies the name of the terraformResourceSchemaInfo to which the list belongs
	TableName string
}

type TableTestParams struct {

	// list Specifies the name of the terraformResourceSchemaInfo to which the list belongs
	TableName string

	ModuleName string

	TableGeneratorName string
}

type SchemaCustomRenderParams struct {
	TableName string
}
