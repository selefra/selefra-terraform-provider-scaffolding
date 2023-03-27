package provider

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"{{.ModuleName}}/selefra_terraform_schema"
)

// example
func GetResource_example() *selefra_terraform_schema.SelefraTerraformResource {
	return &selefra_terraform_schema.SelefraTerraformResource{
		SelefraTableName:      "example",
		TerraformResourceName: "example",
		Description:           "just for example",
		SubTables:             nil,
		ListResourceParamsFunc: func(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]*ResourceRequestParam, *schema.Diagnostics) {
			// TODO
			return nil, nil
		},
	}
}

{{range $key, $value := .ResourceSlice}}
// {{$value.ResourceName}}
func GetResource_{{$value.ResourceName}}() *selefra_terraform_schema.SelefraTerraformResource {
	return &selefra_terraform_schema.SelefraTerraformResource{
		SelefraTableName:      "{{$value.ResourceName}}",
		TerraformResourceName: "{{$value.ResourceName}}",
		Description:           {{$value.Description}},
		SubTables:             nil,
		ListResourceParamsFunc: func(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]*ResourceRequestParam, *schema.Diagnostics) {
			// TODO
			return nil, nil
		},
	}
}

{{end}}


