package generate_selefra_terraform_provider

import (
	"bytes"
	"context"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
	"github.com/yezihack/colorlog"
	"os"
	"path/filepath"
	"text/template"
)

type SchemaGenerator struct {
	config                      *Config
	selefraProviderRenderParams *SelefraProviderRenderParams
}

func NewSchemaGeneratorV2(config *Config, selefraProviderRenderParams *SelefraProviderRenderParams) *SchemaGenerator {
	return &SchemaGenerator{
		config:                      config,
		selefraProviderRenderParams: selefraProviderRenderParams,
	}
}

func (x *SchemaGenerator) Run(ctx context.Context) error {
	t, err := template.New("schema.go").Parse(string(provider_template_v2_generate.SelefraSchemaTemplate))
	if err != nil {
		colorlog.Error("parse schema.go template error: %s", err.Error())
		return err
	}

	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "schema.go", x.selefraProviderRenderParams); err != nil {
		colorlog.Error("render schema.go error: %s", err.Error())
		return err
	}

	schemaGoOutputDirectory := filepath.Join(x.config.Output.Directory, "resources")
	_ = os.MkdirAll(schemaGoOutputDirectory, os.ModePerm)
	schemaGoOutputPath := filepath.Join(schemaGoOutputDirectory, "selefra_schema.go")
	if err := os.WriteFile(schemaGoOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		colorlog.Error("write file %s error: %s", schemaGoOutputPath, err.Error())
		return err
	}
	colorlog.Info("write file %s success", schemaGoOutputPath)
	return nil
}
