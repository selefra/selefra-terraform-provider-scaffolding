package provider

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/terraform/selefra_terraform_schema"
)

// terraform resource: aws_codestarconnections_host
func GetResource_aws_codestarconnections_host() *selefra_terraform_schema.SelefraTerraformResource {
	return &selefra_terraform_schema.SelefraTerraformResource{
		SelefraTableName:      "aws_codestarconnections_host",
		TerraformResourceName: "aws_codestarconnections_host",
		Description:           "",
		SubTables:             nil,
		ListIdsFunc: func(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]string, *schema.Diagnostics) {
			// TODO
			return nil, nil
		},
	}
}
