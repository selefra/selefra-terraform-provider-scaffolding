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
	"github.com/yezihack/colorlog"
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
	colorlog.Info("begin generate client.go...")
	if err := NewClientGenerator(x.config).Run(); err != nil {
		return err
	}
	colorlog.Info("generate client.go success!")

	if err := NewSchemaManagerGenerator(x.config).Run(); err != nil {
		return err
	}

	// gen main.go
	colorlog.Info("begin generate main.go...")
	if err := NewMainGenerator(x.config).Run(); err != nil {
		return err
	}
	colorlog.Info("generate main.go success!")

	// start terraform provider
	colorlog.Info("begin start terraform bridge service...")
	var err error
	err = x.RunTerraformProvider()
	if err != nil {
		return err
	}
	colorlog.Info("terraform bridge service start success!")
	defer func() {
		err := x.bridge.Shutdown()
		if err != nil {
			colorlog.Error("stop terraform bridge service failed: %s", err.Error())
		} else {
			colorlog.Info("stop terraform bridge service success!")
		}
	}()

	// for stat
	totalCount := 0
	successCount := 0
	failedCount := 0
	ignoredCount := 0
	providerGenerator := NewProviderGenerator(x.config)
	// The selefra table structure is then generated from the terraform schema
	x.bridge.GetProvider().ResourcesMap().Range(func(resourceName string, resource shim.Resource) bool {

		totalCount++

		// Supports setting which resources are converted. If not, all resources are converted by default
		if !x.config.IsResourceNeedGenerate(resourceName) {
			colorlog.Info("terraform terraformResourceSchemaInfo %s, is not need generate, so ignored.", resourceName)
			ignoredCount++
			return true
		}

		colorlog.Info("begin generate selefra table from terraform terraformResourceSchemaInfo %s", resourceName)
		err = NewResourceGenerator(x.config, providerGenerator, resourceName, resource).Run()
		if err != nil {
			colorlog.Error("from terraform terraformResourceSchemaInfo %s generate selefra's table failed: %s", resourceName, err.Error())
			failedCount++
			return false
		}
		successCount++
		colorlog.Info("from terraform terraformResourceSchemaInfo %s generate selefra's table success!", resourceName)
		return true
	})
	if err != nil {
		return err
	}
	// print stat information
	colorlog.Info("from terraform terraformResourceSchemaInfo generate selefra's table done, stat: ")
	colorlog.Info("\t\tTotal: %d", totalCount)
	colorlog.Info("\t\tSuccess: %d", successCount)
	colorlog.Info("\t\tIgnored: %d", ignoredCount)
	if failedCount == 0 {
		colorlog.Info("\t\tFailed: %d", ignoredCount)
	} else {
		colorlog.Error("\t\tFailed: %d", ignoredCount)
	}

	// generating a provider
	colorlog.Info("begin generate provider.go...")
	if err := providerGenerator.Run(); err != nil {
		return err
	}
	colorlog.Info("generate provider.go success!")

	// generate go.mod
	colorlog.Info("begin generate go.mod...")
	if err := NewGoModGenerator(x.config).Run(); err != nil {
		return err
	}
	colorlog.Info("generate go.mod success!")

	// TODO Generate Document Information

	colorlog.Info("provider %s generate done", x.config.Selefra.ModuleName)

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
