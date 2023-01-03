package generate_selefra_terraform_provider

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

type SelefraProviderRenderParams struct {
	ProviderName string `json:"provider_name"`
	TableSlice   []*SelefraTableSchemaRenderParams
	// The current table generator depends on which packages need to be imported
	ImportSet  map[string]struct{}
	ModuleName string
}

func (x *SelefraProviderRenderParams) MergeDependencyImports(tableSchemaRenderParams *SelefraTableSchemaRenderParams) {
	if x.ImportSet == nil {
		x.ImportSet = make(map[string]struct{})
	}
	for importString := range tableSchemaRenderParams.ImportSet {
		x.ImportSet[importString] = struct{}{}
	}
}

// ------------------------------------------------ ---------------------------------------------------------------------

// SelefraTableSchemaRenderParams Parameters needed to render the template
type SelefraTableSchemaRenderParams struct {

	// The name of the table generator
	TableSchemaGeneratorName string

	// Name of table
	TableName string

	// Resource name of terraform
	ResourceName string

	// Description of the table
	Description string

	// All the columns in the table
	ColumnSchemaSlice []*SelefraColumnSchemaRenderParams

	// Primary key columns in a table
	PrimaryKeys []string

	// The current table generator depends on which packages need to be imported
	ImportSet map[string]struct{}

	ModuleName string
}

func NewTableSchemaAutoGenRenderParams() *SelefraTableSchemaRenderParams {
	return &SelefraTableSchemaRenderParams{
		ImportSet: make(map[string]struct{}, 0),
	}
}

func (x *SelefraTableSchemaRenderParams) AddDependencyImport(importString string) {
	if x.ImportSet == nil {
		x.ImportSet = make(map[string]struct{})
	}
	x.ImportSet[importString] = struct{}{}
}

func (x *SelefraTableSchemaRenderParams) MergeColumnRenderParamsImport(columnRenderParams *SelefraColumnSchemaRenderParams) {
	for importString := range columnRenderParams.ImportSet {
		x.AddDependencyImport(importString)
	}
}

// ------------------------------------------------ ---------------------------------------------------------------------

// SelefraColumnSchemaRenderParams Represents structural information about a column in a table
type SelefraColumnSchemaRenderParams struct {

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

	ImportSet map[string]struct{}
}

func (x *SelefraColumnSchemaRenderParams) IsPrimaryKey() bool {
	return strings.ToLower(x.ColumnName) == "id"
}

func (x *SelefraColumnSchemaRenderParams) AddDependencyImport(importString string) {
	if x.ImportSet == nil {
		x.ImportSet = make(map[string]struct{})
	}
	x.ImportSet[importString] = struct{}{}
}

// ------------------------------------------------- --------------------------------------------------------------------
