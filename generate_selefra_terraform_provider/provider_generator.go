package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"github.com/yezihack/colorlog"
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

	colorlog.Info("begin generate provider...")
	if err := x.genProvider(); err != nil {
		colorlog.Error("generate provider error: %s", err.Error())
		return err
	}
	colorlog.Info("generate provider success!")

	colorlog.Info("begin generate test provider...")
	if err := x.genTestProvider(); err != nil {
		colorlog.Error("generate test provider error: %s", err.Error())
		return err
	}
	colorlog.Info("generate test provider success!")

	return nil
}

func (x *ProviderGenerator) genProvider() error {

	// step 1. generate provider.go
	t, err := template.New("provider.go").Parse(string(provider_template.ProviderGoTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	renderParams := &ProviderGoRenderParams{
		TerraformProviderShortName:        x.config.Terraform.TerraformProvider.ParseProviderShortName(),
		ModuleName:                        x.config.Selefra.ModuleName,
		TerraformProviderExecuteFileSlice: x.config.Terraform.TerraformProvider.ExecuteFiles,
	}
	if err = t.ExecuteTemplate(&buffer, "provider.go", renderParams); err != nil {
		return err
	}
	providerOutputDirectory := x.config.Output.Directory + "/provider/"
	_ = os.MkdirAll(providerOutputDirectory, os.ModePerm)
	providerOutputFile := providerOutputDirectory + "provider.go"
	if exists, err := PathExists(providerOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", providerOutputFile)
	} else {
		if err := os.WriteFile(providerOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
			return err
		}
		colorlog.Info("generate file %s success", providerOutputFile)
	}

	// step 2. generate provider_test.go
	//t, err = template.New("provider_test.go").Parse(string(provider_template.ProviderTestGoTemplate))
	//if err != nil {
	//	return err
	//}
	//if err = t.ExecuteTemplate(&buffer, "provider_test.go", nil); err != nil {
	//	return err
	//}
	providerTestOutputFile := providerOutputDirectory + "provider_test.go"
	if exists, err := PathExists(providerTestOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", providerTestOutputFile)
	} else {
		if err := os.WriteFile(providerTestOutputFile, []byte(provider_template.ProviderTestGoTemplate), os.ModePerm); err != nil {
			return err
		}
		colorlog.Info("generate file %s success", providerTestOutputFile)
	}

	// step 3. generate tables
	tableGeneratorNameSlice := make([]string, 0)
	for _, tableGeneratorName := range x.TableGeneratorNameSlice {
		tableGeneratorNameSlice = append(tableGeneratorNameSlice, tableGeneratorName)
	}
	t, err = template.New("tables.go").Parse(string(provider_template.ProviderTablesGoTemplate))
	if err != nil {
		return err
	}
	buffer = bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "tables.go", &ProviderTablesGoRenderParams{
		ModuleName:              x.config.Selefra.ModuleName,
		TableGeneratorNameSlice: tableGeneratorNameSlice,
	}); err != nil {
		return err
	}
	providerTablesOutputFile := providerOutputDirectory + "tables.go"
	if exists, err := PathExists(providerTablesOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", providerTablesOutputFile)
	} else {
		if err := os.WriteFile(providerTablesOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
			return err
		}
		colorlog.Info("generate file %s success", providerTablesOutputFile)
	}

	return nil
}

func (x *ProviderGenerator) genTestProvider() error {
	t, err := template.New("test_provider.go").Parse(string(provider_template.TestProviderGoTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "test_provider.go", &ProviderGoRenderParams{
		TerraformProviderShortName:        x.config.Terraform.TerraformProvider.ParseProviderShortName(),
		ModuleName:                        x.config.Selefra.ModuleName,
		TerraformProviderExecuteFileSlice: x.config.Terraform.TerraformProvider.ExecuteFiles,
	}); err != nil {
		return err
	}
	testProviderOutputDirectory := x.config.Output.Directory + "/test_provider/"
	_ = os.MkdirAll(testProviderOutputDirectory, os.ModePerm)
	testProviderOutputFile := testProviderOutputDirectory + "test_provider.go"
	if exists, err := PathExists(testProviderOutputFile); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", testProviderOutputFile)
	} else {
		if err := os.WriteFile(testProviderOutputFile, buffer.Bytes(), os.ModePerm); err != nil {
			return err
		}
		colorlog.Info("generate file %s success", testProviderOutputFile)
	}

	return nil
}

type ProviderGoRenderParams struct {
	TerraformProviderShortName        string
	ModuleName                        string
	TerraformProviderExecuteFileSlice []*provider.TerraformProviderFile
}

type ProviderTablesGoRenderParams struct {
	TableGeneratorNameSlice []string
	ModuleName              string
}
