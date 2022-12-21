package {{.PackageName}}

import (
    "context"
    "github.com/selefra/selefra-provider-sdk/provider/schema"
)

func {{.ResourceName}}_list(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]string, *schema.Diagnostics) {
    // TODO Please implement this list method, documentation reference https://registry.terraform.io/providers/hashicorp/{{.TerraformProviderName}}/latest/docs/resources/{{.ResourceName}}#attributes-reference
    return nil, nil
}