package generate_selefra_terraform_provider

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelefraTerraformProviderInit_Run(t *testing.T) {
	config := &Config{
		Terraform: Terraform{
			TerraformProvider: TerraformProvider{
				RepoUrl: "https://github.com/hashicorp/terraform-provider-aws",
			},
		},
		Output: Output{
			Directory: "./test/",
		},
	}
	files, err := config.Terraform.TerraformProvider.GetTerraformOfficialProviderFiles()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(files))
	err = NewSelefraTerraformProviderInit(config).Run(context.Background())
	assert.Nil(t, err)
}
