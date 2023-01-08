package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
	"github.com/yezihack/colorlog"
	"os"
	"path/filepath"
	"text/template"
)

type SelefraProviderTestGenerator struct {
	config                      *Config
	selefraProviderRenderParams *SelefraProviderRenderParams
}

func NewSelefraProviderTestGenerator(config *Config, selefraProviderRenderParams *SelefraProviderRenderParams) *SelefraProviderTestGenerator {
	return &SelefraProviderTestGenerator{
		config:                      config,
		selefraProviderRenderParams: selefraProviderRenderParams,
	}
}

func (x *SelefraProviderTestGenerator) Run() error {
	t, err := template.New("selefra_provider_test.go").Parse(string(provider_template_v2_generate.SelefraProviderTestTemplate))
	if err != nil {
		colorlog.Error("parse selefra_provider_test.go template error: %s", err.Error())
		return err
	}

	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "selefra_provider_test.go", x.selefraProviderRenderParams); err != nil {
		colorlog.Error("render selefra_provider_test.go error: %s", err.Error())
		return err
	}

	providerGoOutputDirectory := filepath.Join(x.config.Output.Directory, "resources")
	_ = os.MkdirAll(providerGoOutputDirectory, os.ModePerm)
	providerGoOutputPath := filepath.Join(providerGoOutputDirectory, "selefra_provider_test.go")
	if err := os.WriteFile(providerGoOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		colorlog.Error("write file %s error: %s", providerGoOutputPath, err.Error())
		return err
	}
	colorlog.Info("write file %s success", providerGoOutputPath)
	return nil
}
