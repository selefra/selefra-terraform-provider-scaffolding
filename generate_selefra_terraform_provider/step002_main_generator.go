package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
	"os"
	"path/filepath"
	"text/template"
)

type MainGenerator struct {
	config *Config
}

func NewMainGenerator(config *Config) *MainGenerator {
	return &MainGenerator{
		config: config,
	}
}

func (x *MainGenerator) Run() error {

	t, err := template.New("main.go").Parse(string(provider_template_v2_generate.MainTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	renderParams := MainRenderParams{
		ModuleName: x.config.Selefra.ModuleName,
	}
	if err = t.ExecuteTemplate(&buffer, "main.go", renderParams); err != nil {
		return err
	}

	_ = os.MkdirAll(x.config.Output.Directory, os.ModePerm)
	mainFileOutputPath := filepath.Join(x.config.Output.Directory, "main.go")
	if err := os.WriteFile(mainFileOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

type MainRenderParams struct {
	ModuleName string
}
