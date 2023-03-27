package cmd

import (
	"github.com/selefra/selefra-terraform-provider-scaffolding/generate_selefra_terraform_provider"
	"github.com/spf13/cobra"
	"github.com/yezihack/colorlog"
)

func init() {
	rootCmd.AddCommand(generate)
}

var generate = &cobra.Command{
	Use:   "generate",
	Short: "Generate selefra provider from terraform's provider",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		colorlog.Info("begin run generate...")

		var config *generate_selefra_terraform_provider.Config
		var err error
		config, err = generate_selefra_terraform_provider.NewConfigFromLocalJson()
		if err != nil {
			config, err = generate_selefra_terraform_provider.NewConfigFromEnv()
		}
		if err != nil {
			return
		}

		err = generate_selefra_terraform_provider.NewGenerator(config).Run()
		if err != nil {
			colorlog.Error("run generate failed: %s", err.Error())
		} else {
			colorlog.Info("run generate done")
		}

	},
}
