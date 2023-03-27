package generate_selefra_terraform_provider

import (
	"context"
	"encoding/json"
	"fmt"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/terraform/bridge"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/yezihack/colorlog"
	"os"
	"path/filepath"
	"strings"
)

type SchemaIRManager struct {
	config *Config
}

func NewSchemaIRManager(config *Config) *SchemaIRManager {
	return &SchemaIRManager{
		config: config,
	}
}

func (x *SchemaIRManager) GenerateIRAndSave(ctx context.Context) error {
	colorlog.Info("begin generate terraform schema IR...")
	terraformProviderSchemaIR, err := x.GenTerraformProviderSchemaIR(ctx)
	if err != nil {
		colorlog.Error("generate terraform schema IR error: %s", err.Error())
		return err
	}
	colorlog.Info("generate terraform schema IR success, begin save to %s", x.getTerraformSchemaIRSavePath())
	err = x.saveTerraformSchemaIR(terraformProviderSchemaIR)
	if err != nil {
		colorlog.Error("save terraform schema IR to %s error: %s", x.getTerraformSchemaIRSavePath(), err.Error())
		return err
	}
	colorlog.Info("save terraform schema IR to %s success", x.getTerraformSchemaIRSavePath())
	return nil
}

func (x *SchemaIRManager) ReadOrGenerateSchemaIR(ctx context.Context) (*TerraformProviderSchemaIR, error) {
	colorlog.Info("begin read or generate terraform schema IR...")
	ir, err := x.readTerraformSchemaIR()
	if ir == nil {
		colorlog.Info("not found before schema.json, so generate it...")
		err := x.GenerateIRAndSave(ctx)
		if err != nil {
			colorlog.Error("generate terraform schema IR error: %s", err.Error())
			return nil, err
		}
	}
	schemaIR, err := x.readTerraformSchemaIR()
	if err != nil {
		colorlog.Error("read terraform schema IR error: %s", err.Error())
		return nil, err
	}
	colorlog.Info("read terraform schema IR error success")
	return schemaIR, nil
}

func (x *SchemaIRManager) GenTerraformProviderSchemaIR(ctx context.Context) (*TerraformProviderSchemaIR, error) {
	colorlog.Info("begin start terraform provider bridge for %s ...", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
	terraformProviderBridge, err := x.RunTerraformProvider(ctx)
	if err != nil {
		colorlog.Error("start terraform provider bridge for %s error: %s", x.config.Terraform.TerraformProvider.GetOrParseProviderName(), err.Error())
		return nil, err
	}
	colorlog.Info("start terraform provider bridge %s success", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
	defer func() {
		err := terraformProviderBridge.Shutdown()
		if err != nil {
			colorlog.Error("terraform provider bridge %s shutdown failed: %s", x.config.Terraform.TerraformProvider.GetOrParseProviderName(), err.Error())
		} else {
			colorlog.Info("terraform provider bridge %s shutdown success", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
		}
	}()
	return FromTerraformProviderSchema(x.config.Terraform.TerraformProvider.GetOrParseProviderName(), terraformProviderBridge.GetProvider(), x.config), nil
}

func (x *SchemaIRManager) RunTerraformProvider(ctx context.Context) (*bridge.TerraformBridge, error) {
	providerExecFileSaveDirectory := filepath.Join(x.config.Output.Directory, "/bin/", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
	colorlog.Info("begin download provider %s's exec file to %s", x.config.Terraform.TerraformProvider.GetOrParseProviderName(), providerExecFileSaveDirectory)
	providerExecFilePath, err := provider.NewProviderDownloader(x.config.Terraform.TerraformProvider.ExecuteFiles).Download(providerExecFileSaveDirectory)
	if err != nil {
		colorlog.Error("download provider %s's exec file failed: %s", x.config.Terraform.TerraformProvider.GetOrParseProviderName(), err.Error())
		return nil, err
	}
	terraformProviderBridge := bridge.NewTerraformBridge(providerExecFilePath)
	// Some providers need to configure parameters at startup
	providerConfig := make(map[string]any, 0)
	if x.config.Terraform.TerraformProvider.Config != "" {
		err := json.Unmarshal([]byte(x.config.Terraform.TerraformProvider.Config), &providerConfig)
		if err != nil {
			colorlog.Error("json unmarshal provider config error, raw = %s, err msg = %s", x.config.Terraform.TerraformProvider.Config, err.Error())
			return nil, fmt.Errorf("json unmarshal terraform provider config error: %+v", err)
		}
	}
	colorlog.Info("begin run bridge for provider %s...", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
	err = terraformProviderBridge.StartBridge(ctx, providerConfig)
	if err != nil {
		colorlog.Error("run bridge for provider %s failed: %s", x.config.Terraform.TerraformProvider.GetOrParseProviderName(), err.Error())
		return nil, err
	}
	colorlog.Info("run bridge for provider %s success", x.config.Terraform.TerraformProvider.GetOrParseProviderName())
	return terraformProviderBridge, nil
}

func (x *SchemaIRManager) getTerraformSchemaIRSavePath() string {
	schemaJsonOutputDirectory := filepath.Join(x.config.Output.Directory, "/provider")
	_ = os.MkdirAll(schemaJsonOutputDirectory, os.ModePerm)
	return filepath.Join(schemaJsonOutputDirectory, "/schema.json")
}

func (x *SchemaIRManager) saveTerraformSchemaIR(terraformProviderSchemaIR *TerraformProviderSchemaIR) error {
	marshal, err := json.Marshal(terraformProviderSchemaIR)
	if err != nil {
		colorlog.Error("save terraform schema IR failed: %s", err.Error())
		return err
	}
	if err := os.WriteFile(x.getTerraformSchemaIRSavePath(), marshal, os.ModePerm); err != nil {
		colorlog.Error("save terraform schema IR failed: %s", err.Error())
		return err
	}
	return nil
}

func (x *SchemaIRManager) readTerraformSchemaIR() (*TerraformProviderSchemaIR, error) {
	schemaBytes, err := os.ReadFile(x.getTerraformSchemaIRSavePath())
	if err != nil {
		colorlog.Error("read terraform schema IR failed: %s", err.Error())
		return nil, err
	}
	terraformProviderSchemaIR := &TerraformProviderSchemaIR{}
	err = json.Unmarshal(schemaBytes, &terraformProviderSchemaIR)
	if err != nil {
		colorlog.Error("read terraform schema IR failed: %s", err.Error())
		return nil, err
	}
	return terraformProviderSchemaIR, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

type TerraformProviderSchemaIR struct {
	ProviderName string                       `json:"provider_name"`
	Resources    []*TerraformResourceSchemaIR `json:"resources"`
}

func FromTerraformProviderSchema(terraformProviderName string, provider shim.Provider, config *Config) *TerraformProviderSchemaIR {
	terraformProviderSchemaIR := &TerraformProviderSchemaIR{
		ProviderName: terraformProviderName,
	}
	provider.ResourcesMap().Range(func(terraformResourceName string, terraformResourceSchema shim.Resource) bool {

		if !config.IsResourceNeedGenerate(terraformResourceName) {
			colorlog.Info("terraform resource %s, do not need generate, so ignored", terraformResourceName)
			return true
		}

		resourceSchemaIR := FromTerraformResourceSchema(terraformResourceName, terraformResourceSchema, config)
		if resourceSchemaIR == nil {
			return true
		}
		terraformProviderSchemaIR.Resources = append(terraformProviderSchemaIR.Resources, resourceSchemaIR)
		return true
	})
	return terraformProviderSchemaIR
}

func (x *TerraformProviderSchemaIR) ToSelefraProviderRenderParams(selefraModuleName string) *SelefraProviderRenderParams {
	providerRenderParams := &SelefraProviderRenderParams{
		ProviderName: x.ProviderName,
		ModuleName:   selefraModuleName,
	}
	for _, resourceSchemeIR := range x.Resources {
		selefraTableRender := resourceSchemeIR.ToSelefraTableRenderParams(selefraModuleName)
		if selefraTableRender == nil {
			continue
		}
		providerRenderParams.TableSlice = append(providerRenderParams.TableSlice, selefraTableRender)
		providerRenderParams.MergeDependencyImports(selefraTableRender)
	}
	return providerRenderParams
}

// ------------------------------------------------- --------------------------------------------------------------------

type TerraformResourceSchemaIR struct {
	ResourceName string                     `json:"resource_name"`
	Description  string                     `json:"description"`
	Columns      []*TerraformColumnSchemaIR `json:"columns"`
}

func FromTerraformResourceSchema(terraformResourceName string, terraformResourceSchema shim.Resource, config *Config) *TerraformResourceSchemaIR {
	resourceSchema := &TerraformResourceSchemaIR{
		ResourceName: terraformResourceName,
		Description:  "",
	}
	isNeedGenerate := false
	terraformResourceSchema.Schema().Range(func(terraformColumnName string, terraformColumnSchema shim.Schema) bool {

		//if !config.IsResourceNeedGenerate(terraformColumnName) {
		//	colorlog.Info("terraform resource %s, do not need generate, so ignored", terraformColumnName)
		//	return true
		//}
		columnSchema := FromTerraformColumnSchema(terraformColumnName, terraformColumnSchema)
		if columnSchema == nil {
			return true
		}
		resourceSchema.Columns = append(resourceSchema.Columns, columnSchema)
		if columnSchema.IsID() {
			isNeedGenerate = true
		}
		return true
	})
	//if !hasIdColumn {
	//	colorlog.Error("terraform resource %s do not have id column, so ignored", terraformResourceName)
	//	return nil
	//}

	if !isNeedGenerate {
		return nil
	}

	return resourceSchema
}

func (x *TerraformResourceSchemaIR) ToSelefraTableRenderParams(selefraModuleName string) *SelefraTableSchemaRenderParams {
	tableParams := &SelefraTableSchemaRenderParams{
		TableSchemaGeneratorName: x.BuildTableSchemaGeneratorName(),
		TableName:                x.ResourceName,
		Description:              processDescription(x.Description),
		PrimaryKeys:              []string{"id"},
		ModuleName:               selefraModuleName,
	}

	hasIdColumn := false
	for _, column := range x.Columns {
		renderParams := column.ToSelefraSchemaRenderParams()
		tableParams.ColumnSchemaSlice = append(tableParams.ColumnSchemaSlice, renderParams)
		if column.IsID() {
			hasIdColumn = true
		}
		tableParams.MergeColumnRenderParamsImport(renderParams)
	}
	if !hasIdColumn {
		colorlog.Error("terraform resource %s do not have id column, so ignored", x.ResourceName)
		return nil
	}

	// Add an additional column to store the original response data
	tableParams.ColumnSchemaSlice = append(tableParams.ColumnSchemaSlice, &SelefraColumnSchemaRenderParams{
		ColumnName:                "selefra_terraform_original_result",
		Description:               "`save terraform original result for compatibility`",
		ColumnTypeCodeString:      "schema.ColumnTypeJSON",
		ExtractorInlineCodeString: "column_value_extractor.TerraformRawDataColumnValueExtractor()",
	})

	return tableParams
}

func (x *TerraformResourceSchemaIR) BuildTableSchemaGeneratorName() string {
	s := strings.Replace(x.ResourceName, "_", " ", -1)
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1) + "SchemaGenerator"
}

// ------------------------------------------------- --------------------------------------------------------------------

type TerraformColumnSchemaIR struct {
	ColumnName  string            `json:"column_name"`
	ColumnType  schema.ColumnType `json:"column_type"`
	Description string            `json:"description"`
}

// FromTerraformColumnSchema Generates intermediate structure information from the column structure of the terraform
func FromTerraformColumnSchema(terraformColumnName string, terraformColumnSchema shim.Schema) *TerraformColumnSchemaIR {
	columnSchema := &TerraformColumnSchemaIR{
		ColumnName:  terraformColumnName,
		Description: terraformColumnSchema.Description(),
	}

	// column's type & column value extractor
	switch terraformColumnSchema.Type() {
	case shim.TypeBool:
		columnSchema.ColumnType = schema.ColumnTypeBool
	case shim.TypeInt:
		columnSchema.ColumnType = schema.ColumnTypeBigInt
	case shim.TypeFloat:
		columnSchema.ColumnType = schema.ColumnTypeFloat
	case shim.TypeString:
		columnSchema.ColumnType = schema.ColumnTypeString
	case shim.TypeList, shim.TypeMap, shim.TypeSet:
		// All are converted to JSON
		columnSchema.ColumnType = schema.ColumnTypeJSON
	case shim.TypeInvalid:
		columnSchema.ColumnType = schema.ColumnTypeNotAssign
	}

	return columnSchema
}

func (x *TerraformColumnSchemaIR) ToSelefraSchemaRenderParams() *SelefraColumnSchemaRenderParams {

	selefraColumnRenderParams := &SelefraColumnSchemaRenderParams{
		ColumnName:  x.ColumnName,
		Description: processDescription(x.Description),
	}

	// column's type & column value extractor
	switch x.ColumnType {
	case schema.ColumnTypeBool:
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeBool"
	case schema.ColumnTypeBigInt:
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeBigInt"
	case schema.ColumnTypeFloat:
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeFloat"
	case schema.ColumnTypeString:
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeString"
	case schema.ColumnTypeJSON:
		// All are converted to JSON
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeJSON"
		selefraColumnRenderParams.ExtractorInlineCodeString = fmt.Sprintf("column_value_extractor.TerraformRawDataColumnValueExtractor(\"%s\")", x.ColumnName)
		selefraColumnRenderParams.AddDependencyImport("github.com/selefra/selefra-provider-sdk/terraform/column_value_extractor")
	case schema.ColumnTypeNotAssign:
		selefraColumnRenderParams.ColumnTypeCodeString = "schema.ColumnTypeInvalid"
	}

	return selefraColumnRenderParams
}

func (x *TerraformColumnSchemaIR) IsID() bool {
	return strings.ToLower(x.ColumnName) == "id"
}

// ------------------------------------------------- --------------------------------------------------------------------
