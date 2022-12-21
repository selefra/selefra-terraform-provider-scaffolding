package generate_selefra_terraform_provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	"github.com/selefra/selefra-provider-sdk/terraform/bridge"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/spf13/viper"
	"os"
)

type Generator struct {
	config *Config

	bridge *bridge.TerraformBridge
}

func NewGenerateTerraformProvider(configFilePath string) (*Generator, error) {

	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	viperConfig := viper.New()
	viperConfig.SetConfigType("yaml")
	err = viperConfig.ReadConfig(bytes.NewReader(configBytes))
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = viperConfig.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	if err := checkConfig(config); err != nil {
		return nil, err
	}

	return &Generator{
		config: config,
	}, nil
}

func checkConfig(config *Config) error {

	if config.Selefra.ModuleName == "" {
		return fmt.Errorf("selefra.module-name must set")
	}

	if config.Output.Directory == "" {
		return fmt.Errorf("output.directory must set")
	}

	if config.Terraform.TerraformProvider.ParseProviderName() == "" {
		return fmt.Errorf("can not parse provider name from : %s", config.Terraform.TerraformProvider.RepoUrl)
	}

	// It is the official provider
	if b, _ := config.Terraform.TerraformProvider.IsTerraformOfficialProvider(); b {
		files, err := config.Terraform.TerraformProvider.GetTerraformOfficialProviderFiles()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("You have specified an official provider, but I cannot automatically parse the corresponding provider file. Please make the provider file manually")
		}
	}

	// It's a provider on github
	if b, _ := config.Terraform.TerraformProvider.IsGithubRepo(); len(config.Terraform.TerraformProvider.ExecuteFiles) == 0 && b {
		files, err := config.Terraform.TerraformProvider.RequestGithubReleaseFiles()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("You specified a provider hosted on Github, but I cannot automatically parse the corresponding provider file. Please specify the provider file manually")
		}
	}

	if len(config.Terraform.TerraformProvider.ExecuteFiles) == 0 {
		return fmt.Errorf("It's a provider on github")
	}

	return nil
}

func (x *Generator) Run() error {

	// step 001. first generate client code
	if err := NewClientGenerator(x.config).Run(); err != nil {
		return err
	}

	// start terraform provider
	var err error
	err = x.RunTerraformProvider()
	if err != nil {
		return err
	}

	providerGenerator := NewProviderGenerator(x.config)

	// The selefra table structure is then generated from the terraform schema
	x.bridge.GetProvider().ResourcesMap().Range(func(resourceName string, resource shim.Resource) bool {

		// Supports setting which resources are converted. If not, all resources are converted by default
		if !x.config.IsResourceNeedGenerate(resourceName) {
			return true
		}

		err = NewResourceGenerator(x.config, providerGenerator, resourceName, resource).Run()
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	// Generating a provider
	if err := providerGenerator.Run(); err != nil {
		return err
	}

	// generate go.mod
	if err := NewGoModGenerator(x.config).Run(); err != nil {
		return err
	}

	// TODO Generate Document Information

	return nil
}

func (x *Generator) RunTerraformProvider() error {
	providerExecFilePath, err := provider.NewProviderDownloader(x.config.Terraform.TerraformProvider.ExecuteFiles).Download("./" + x.config.Terraform.TerraformProvider.ParseProviderName())
	if err != nil {
		return err
	}
	x.bridge = bridge.NewTerraformBridge(providerExecFilePath)
	// Some providers need to configure parameters at startup
	providerConfig := make(map[string]any, 0)
	if x.config.Terraform.TerraformProvider.Config != "" {
		err := json.Unmarshal([]byte(x.config.Terraform.TerraformProvider.Config), &providerConfig)
		if err != nil {
			return fmt.Errorf("json unmarshal terraform provider config error: %+v", err)
		}
	}
	err = x.bridge.StartBridge(context.Background(), providerConfig)
	if err != nil {
		return err
	}
	return nil
}
