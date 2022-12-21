package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"os"
	"text/template"
)

type ProviderGenerator struct {
	config                  *Config
	TableGeneratorNameSlice []string
}

func NewProviderGenerator(config *Config) *ProviderGenerator {
	return &ProviderGenerator{
		config: config,
	}
}

func (x *ProviderGenerator) Run() error {

	if err := x.genProvider(); err != nil {
		return err
	}

	if err := x.genTestProvider(); err != nil {
		return err
	}

	return nil
}

func (x *ProviderGenerator) genProvider() error {
	t, err := template.New("provider").Parse(string(provider_template.ProviderTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "provider", &RenderProviderParams{
		TerraformProviderName:             x.config.Terraform.TerraformProvider.ParseShortProviderName(),
		GoModuleName:                      x.config.Selefra.ModuleName,
		TerraformProviderExecuteFileSlice: x.config.Terraform.TerraformProvider.ExecuteFiles,
	}); err != nil {
		return err
	}
	providerOutputDirectory := x.config.Output.Directory + "/provider/"
	_ = os.MkdirAll(providerOutputDirectory, os.ModePerm)
	providerOutputFile := providerOutputDirectory + "provider.go"
	if err := os.WriteFile(providerOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// tables
	tableGeneratorNameSlice := make([]string, 0)
	for _, tableGeneratorName := range x.TableGeneratorNameSlice {
		tableGeneratorNameSlice = append(tableGeneratorNameSlice, tableGeneratorName)
	}
	t, err = template.New("table").Parse(string(provider_template.ProviderTablesTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "table", &RenderProviderTableParams{
		GoModuleName:            x.config.Selefra.ModuleName,
		TableGeneratorNameSlice: tableGeneratorNameSlice,
	}); err != nil {
		return err
	}
	providerTablesOutputFile := providerOutputDirectory + "tables.go"
	if err := os.WriteFile(providerTablesOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (x *ProviderGenerator) genTestProvider() error {
	t, err := template.New("test-provider").Parse(string(provider_template.TestProvider))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "test-provider", &RenderProviderParams{
		TerraformProviderName:             x.config.Terraform.TerraformProvider.ParseShortProviderName(),
		GoModuleName:                      x.config.Selefra.ModuleName,
		TerraformProviderExecuteFileSlice: x.config.Terraform.TerraformProvider.ExecuteFiles,
	}); err != nil {
		return err
	}
	providerOutputDirectory := x.config.Output.Directory + "/test_provider/"
	_ = os.MkdirAll(providerOutputDirectory, os.ModePerm)
	providerOutputFile := providerOutputDirectory + "test_provider.go"
	if err := os.WriteFile(providerOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

type RenderProviderParams struct {
	TerraformProviderName             string
	GoModuleName                      string
	TerraformProviderExecuteFileSlice []*provider.TerraformProviderFile
}

type RenderProviderTableParams struct {
	ImportSlice             []string
	TableGeneratorNameSlice []string
	GoModuleName            string
}
