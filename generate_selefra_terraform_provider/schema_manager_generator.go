package generate_selefra_terraform_provider

import (
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"os"
)

type SchemaManagerGenerator struct {
	config *Config
}

func NewSchemaManagerGenerator(config *Config) *SchemaManagerGenerator {
	return &SchemaManagerGenerator{
		config: config,
	}
}

func (x *SchemaManagerGenerator) Run() error {
	clientDirectory := x.config.Output.Directory + "/schema_manager"
	_ = os.MkdirAll(clientDirectory, os.ModePerm)
	clientFilePath := clientDirectory + "/schema_manager.go"
	if err := os.WriteFile(clientFilePath, []byte(provider_template.SchemaManagerGoTemplate), os.ModePerm); err != nil {
		return err
	}
	return nil
}
