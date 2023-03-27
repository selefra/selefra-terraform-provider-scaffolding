package generate_selefra_terraform_provider

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/terraform/bridge"
)

type Generator struct {
	config *Config
	bridge *bridge.TerraformBridge
}

func NewGenerator(config *Config) *Generator {
	return &Generator{
		config: config,
	}
}

func (x *Generator) Run() error {

	terraformSchemaIR, err := NewSchemaIRManager(x.config).ReadOrGenerateSchemaIR(context.Background())
	if err != nil {
		return err
	}

	if err := NewCopyProvider(x.config).Run(); err != nil {
		return err
	}

	selefraProviderRenderParams := terraformSchemaIR.ToSelefraProviderRenderParams(x.config.Selefra.ModuleName)
	if err := NewSchemaGeneratorV2(x.config, selefraProviderRenderParams).Run(context.Background()); err != nil {
		return err
	}

	if err := NewProviderGenerator(x.config, selefraProviderRenderParams).Run(); err != nil {
		return err
	}

	if err := NewSelefraProviderTestGenerator(x.config, selefraProviderRenderParams).Run(); err != nil {
		return err
	}

	if err := NewMainGenerator(x.config).Run(); err != nil {
		return err
	}

	//if err := NewGoModGenerator(x.config).Run(); err != nil {
	//	return err
	//}

	// ------------------------------------------------- --------------------------------------------------------------------

	//// step 001. first generate client code
	//colorlog.Info("begin generate client.go...")
	//if err := NewClientGenerator(x.config).Run(); err != nil {
	//	return err
	//}
	//colorlog.Info("generate client.go success!")
	//
	//if err := NewSchemaManagerGenerator(x.config).Run(); err != nil {
	//	return err
	//}
	//
	//// gen main.go
	//colorlog.Info("begin generate main.go...")
	//if err := NewMainGenerator(x.config).Run(); err != nil {
	//	return err
	//}
	//colorlog.Info("generate main.go success!")
	//
	//// start terraform provider
	//colorlog.Info("begin start terraform bridge service...")
	//var err error
	//err = x.RunTerraformProvider()
	//if err != nil {
	//	return err
	//}
	//colorlog.Info("terraform bridge service start success!")
	//defer func() {
	//	err := x.bridge.Shutdown()
	//	if err != nil {
	//		colorlog.Error("stop terraform bridge service failed: %s", err.Error())
	//	} else {
	//		colorlog.Info("stop terraform bridge service success!")
	//	}
	//}()
	//
	//// for stat
	//totalCount := 0
	//successCount := 0
	//failedCount := 0
	//ignoredCount := 0
	//providerGenerator := NewProviderGenerator(x.config)
	//// The selefra table structure is then generated from the terraform schema
	//x.bridge.GetProvider().ResourcesMap().Range(func(resourceName string, resource shim.Resource) bool {
	//
	//	totalCount++
	//
	//	// Supports setting which resources are converted. If not, all resources are converted by default
	//	if !x.config.IsResourceNeedGenerate(resourceName) {
	//		colorlog.Info("terraform terraformResourceSchemaInfo %s, is not need generate, so ignored.", resourceName)
	//		ignoredCount++
	//		return true
	//	}
	//
	//	colorlog.Info("begin generate selefra table from terraform terraformResourceSchemaInfo %s", resourceName)
	//	err = NewResourceGenerator(x.config, providerGenerator, resourceName, resource).Run()
	//	if err != nil {
	//		colorlog.Error("from terraform terraformResourceSchemaInfo %s generate selefra's table failed: %s", resourceName, err.Error())
	//		failedCount++
	//		return false
	//	}
	//	successCount++
	//	colorlog.Info("from terraform terraformResourceSchemaInfo %s generate selefra's table success!", resourceName)
	//	return true
	//})
	//if err != nil {
	//	return err
	//}
	//// print stat information
	//colorlog.Info("from terraform terraformResourceSchemaInfo generate selefra's table done, stat: ")
	//colorlog.Info("\t\tTotal: %d", totalCount)
	//colorlog.Info("\t\tSuccess: %d", successCount)
	//colorlog.Info("\t\tIgnored: %d", ignoredCount)
	//if failedCount == 0 {
	//	colorlog.Info("\t\tFailed: %d", ignoredCount)
	//} else {
	//	colorlog.Error("\t\tFailed: %d", ignoredCount)
	//}
	//
	//// generating a provider
	//colorlog.Info("begin generate provider.go...")
	//if err := providerGenerator.Run(); err != nil {
	//	return err
	//}
	//colorlog.Info("generate provider.go success!")
	//
	//// generate go.mod
	//colorlog.Info("begin generate go.mod...")
	//if err := NewGoModGenerator(x.config).Run(); err != nil {
	//	return err
	//}
	//colorlog.Info("generate go.mod success!")
	//
	//// TODO Generate Document Information
	//
	//colorlog.Info("provider %s generate done", x.config.Selefra.ModuleName)

	return nil
}
