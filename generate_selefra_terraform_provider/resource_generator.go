package generate_selefra_terraform_provider

import (
	"bytes"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"os"
	"strings"
	"text/template"
)

// ResourceGenerator Generate resources related to resource
type ResourceGenerator struct {

	// The configuration information is generated
	config *Config

	// the provider to which the current resource belongs
	provider *ProviderGenerator

	// The name of the resource currently being processed
	resourceName string
	resource     shim.Resource

	// Parameters when the table is generated from the resource
	params *TableGeneratorParams
}

// NewResourceGenerator Create a resource to table generator
func NewResourceGenerator(config *Config, provider *ProviderGenerator, resourceName string, resource shim.Resource) *ResourceGenerator {
	return &ResourceGenerator{
		config:       config,
		provider:     provider,
		resourceName: resourceName,
		resource:     resource,
		params:       NewTableGeneratorParams(),
	}
}

func (x *ResourceGenerator) Run() error {

	// Generate table names and the like
	x.params.ResourceName = x.resourceName
	x.params.PackageName = "tables"
	x.params.TableName = x.resourceName
	x.params.TableSchemaGeneratorName = snakeToCamel(x.resourceName) + "Generator"
	x.params.GoModuleName = x.config.Selefra.ModuleName

	// Generate the columns of the table
	if isHaveIdColumn, err := x.GenColumns(); err != nil || !isHaveIdColumn {
		return err
	}

	// Rendering template
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

func (x *ResourceGenerator) RenderAndSave() error {

	// First render the table structure
	t, err := template.New("resource").Parse(string(provider_template.ResourceTableSchemaTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "resource", x.params); err != nil {
		return err
	}

	resourceOutputDirectory := x.config.Output.Directory + "/tables/"
	_ = os.MkdirAll(resourceOutputDirectory, os.ModePerm)
	resourceOutputFile := resourceOutputDirectory + x.resourceName + "__schema.go"
	if err := os.WriteFile(resourceOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// And then the list function
	resourceListFuncOutputFile := resourceOutputDirectory + x.resourceName + "__list.go"
	t, err = template.New("table-list").Parse(string(provider_template.ResourceTableListTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "table-list", &ResourceListParams{
		PackageName:           "tables",
		TerraformProviderName: x.config.Terraform.TerraformProvider.ParseProviderName(),
		ResourceName:          x.resourceName,
	}); err != nil {
		return err
	}
	if err := os.WriteFile(resourceListFuncOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// And table test file
	resourceTestOutputFile := resourceOutputDirectory + x.resourceName + "__test.go"
	t, err = template.New("table-test").Parse(string(provider_template.TableTest))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "table-test", &ResourceTestParams{
		ResourceName:       x.resourceName,
		GoModuleUrl:        x.config.Selefra.ModuleName,
		TableGeneratorName: x.params.TableSchemaGeneratorName,
	}); err != nil {
		return err
	}
	if err := os.WriteFile(resourceTestOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// Rendering is successful. Report yourself
	x.provider.TableGeneratorNameSlice = append(x.provider.TableGeneratorNameSlice, x.params.TableSchemaGeneratorName)

	return nil
}

// GenColumns Generate the columns corresponding to the table
func (x *ResourceGenerator) GenColumns() (bool, error) {
	haveIdColumn := false
	x.resource.Schema().Range(func(columnName string, value shim.Schema) bool {
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

// TableGeneratorParams Parameters needed to render the template
type TableGeneratorParams struct {

	// The name of the package to which it belongs
	PackageName string

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

	ResourceName string

	GoModuleName string
}

func NewTableGeneratorParams() *TableGeneratorParams {
	return &TableGeneratorParams{
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

// ResourceListParams Parameter to generate a List resource
type ResourceListParams struct {

	// Which package is the generated result placed under
	PackageName string

	// Name of the terraform provider
	TerraformProviderName string

	// list Specifies the name of the resource to which the list belongs
	ResourceName string
}

type ResourceTestParams struct {

	// list Specifies the name of the resource to which the list belongs
	ResourceName string

	GoModuleUrl string

	TableGeneratorName string
}
