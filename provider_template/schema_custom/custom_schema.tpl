package schema_custom

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
)

// Note: This file is only generated when it does not exist. If the file already exists, it will not be generated. If you want to regenerate this file, please manually delete it first

func CustomDescription_{{.TableName}}() string {
    // You can customize the table description. If you do not specify the terraform table description, the default is TerraForm
	return ""
}

func CustomOptions_{{.TableName}}(options *schema.TableOptions) {
    // If you need to set the options for the table yourself, you can do so here
}

func Version_{{.TableName}}() uint64 {
	return 0
}

func Expand_{{.TableName}}() func(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask) []*schema.ClientTaskContext {
    // If your provider involves multiple accounts, or multiple regions, specify the logic to extend to multiple tasks here
	return nil
}

func CustomExtraColumns_{{.TableName}}() []*schema.Column {
    // You can customize some columns based on the original terraform, if necessary
	return nil
}

func SubTables_{{.TableName}}() []string {
    // This is where you specify the association relationships between tables
	return nil
}
