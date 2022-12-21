package generate_selefra_terraform_provider

import "testing"

func TestGenerateProviderExecuteFiles(t *testing.T) {
	targetUrl := "https://releases.hashicorp.com/terraform-provider-aws/4.47.0/"
	GenerateProviderExecuteFiles(targetUrl)
}
