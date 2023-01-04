package cmd

import (
	"context"
	"github.com/selefra/selefra-terraform-provider-scaffolding/generate_selefra_terraform_provider"
	"github.com/spf13/cobra"
	"github.com/yezihack/colorlog"
)

func init() {
	rootCmd.AddCommand(initSelefraTerraformProvider)
}

var initSelefraTerraformProvider = &cobra.Command{
	Use:   "init",
	Short: "init selefra terraform provider project template",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		colorlog.Info("begin exec init...")

		config, err := generate_selefra_terraform_provider.NewConfigFromEnv()
		if err != nil {
			return
		}

		err = generate_selefra_terraform_provider.NewSelefraTerraformProviderInit(config).Run(context.Background())
		if err != nil {
			colorlog.Error("exec init failed: %s", err.Error())
		} else {
			colorlog.Info("exec init done")
		}

	},
}
