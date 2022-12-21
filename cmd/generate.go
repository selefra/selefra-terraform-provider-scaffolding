package cmd

import (
	"github.com/selefra/selefra-terraform-provider-scaffolding/generate_selefra_terraform_provider"
	"github.com/spf13/cobra"
	"log"
)

var configFilePath string

func init() {
	generate.Flags().StringVarP(&configFilePath, "config", "c", "", "yaml config file path")
	rootCmd.AddCommand(generate)
}

var generate = &cobra.Command{
	Use:   "generate",
	Short: "Generate selefra provider from terraform's provider",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		generator, err := generate_selefra_terraform_provider.NewGenerateTerraformProvider(configFilePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = generator.Run()
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Done")
		}
	},
}
