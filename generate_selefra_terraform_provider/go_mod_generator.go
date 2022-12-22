package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"os"
	"text/template"
)

// GoModGenerator Used to render go.mod
type GoModGenerator struct {
	config *Config
}

func NewGoModGenerator(config *Config) *GoModGenerator {
	return &GoModGenerator{
		config: config,
	}
}

func (x *GoModGenerator) Run() error {
	return x.Render()
}

func (x *GoModGenerator) Render() error {
	t, err := template.New("go.mod").Parse(string(provider_template.GoModTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	params := &GoModRenderParams{
		ModuleName: x.config.Selefra.ModuleName,
	}
	if err = t.ExecuteTemplate(&buffer, "go.mod", params); err != nil {
		return err
	}

	_ = os.MkdirAll(x.config.Output.Directory, os.ModePerm)
	goModOutputPath := x.config.Output.Directory + "/go.mod"
	if err := os.WriteFile(goModOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

type GoModRenderParams struct {
	ModuleName string
}
