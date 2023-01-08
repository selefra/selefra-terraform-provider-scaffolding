package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
	"github.com/yezihack/colorlog"
	"os"
	"path/filepath"
	"text/template"
)

type ProviderGenerator struct {
	config                      *Config
	selefraProviderRenderParams *SelefraProviderRenderParams
}

func NewProviderGenerator(config *Config, selefraProviderRenderParams *SelefraProviderRenderParams) *ProviderGenerator {
	return &ProviderGenerator{
		config:                      config,
		selefraProviderRenderParams: selefraProviderRenderParams,
	}
}

func (x *ProviderGenerator) Run() error {
	t, err := template.New("provider.go").Parse(string(provider_template_v2_generate.SelefraProviderTemplate))
	if err != nil {
		colorlog.Error("parse provider.go template error: %s", err.Error())
		return err
	}

	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "provider.go", x.selefraProviderRenderParams); err != nil {
		colorlog.Error("render provider.go error: %s", err.Error())
		return err
	}

	providerGoOutputDirectory := filepath.Join(x.config.Output.Directory, "resources")
	_ = os.MkdirAll(providerGoOutputDirectory, os.ModePerm)
	providerGoOutputPath := filepath.Join(providerGoOutputDirectory, "selefra_provider.go")
	if err := os.WriteFile(providerGoOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		colorlog.Error("write file %s error: %s", providerGoOutputPath, err.Error())
		return err
	}
	colorlog.Info("write file %s success", providerGoOutputPath)
	return nil
}
