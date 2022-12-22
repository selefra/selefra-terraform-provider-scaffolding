package generate_selefra_terraform_provider

import (
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template"
	"github.com/yezihack/colorlog"
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
	clientOutputFilePath := clientDirectory + "/client.go"
	if exists, err := PathExists(clientOutputFilePath); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", clientOutputFilePath)
		return nil
	}
	if err := os.WriteFile(clientOutputFilePath, []byte(provider_template.ClientTemplate), os.ModePerm); err != nil {
		return err
	}

	colorlog.Info("generate file %s success", clientOutputFilePath)

	return nil
}
