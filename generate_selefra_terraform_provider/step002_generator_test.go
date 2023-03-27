package generate_selefra_terraform_provider

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerator_Run(t *testing.T) {

	// old test case
	//configFilePath := "./test/config-test.yml"
	//provider, err := NewGenerateTerraformProvider(configFilePath)
	//assert.Nil(t, err)
	//assert.NotNil(t, provider)
	//err = provider.Run()
	//assert.Nil(t, err)

	terraformProviderUrl := "https://github.com/hashicorp/terraform-provider-aws"
	config, err := NewConfigFromTerraformProviderRepoUrl(terraformProviderUrl)
	assert.Nil(t, err)
	assert.NotNil(t, config)

	config.Output.Directory = "./test"
	err = NewGenerator(config).Run()
	assert.Nil(t, err)

}
