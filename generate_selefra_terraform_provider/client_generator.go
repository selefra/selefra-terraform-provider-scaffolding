package generate_selefra_terraform_provider

import (
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"os"
)

// ClientGenerator Used to generate client
type ClientGenerator struct {
	config *Config
}

func NewClientGenerator(config *Config) *ClientGenerator {
	return &ClientGenerator{
		config: config,
	}
}

func (x *ClientGenerator) Run() error {
	clientDirectory := x.config.Output.Directory + "/client"
	_ = os.MkdirAll(clientDirectory, os.ModePerm)
	clientFilePath := clientDirectory + "/client.go"
	if err := os.WriteFile(clientFilePath, []byte(provider_template.ClientTemplate), os.ModePerm); err != nil {
		return err
	}
	return nil
}
