package schema_custom

import (
    "context"
    "github.com/selefra/selefra-provider-sdk/provider/schema"
)

func ListIds_{{.TableName}}(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]string, *schema.Diagnostics) {
    // TODO Please implement this list method, documentation reference https://registry.terraform.io/providers/hashicorp/{{.TerraformProviderShortName}}/latest/docs/resources/{{.TableName}}#attributes-reference
    return nil, nil
}