package cmd

//import (
//	"github.com/selefra/selefra-terraform-provider-scaffolding/generate_selefra_terraform_provider"
//	"github.com/spf13/cobra"
//	"github.com/yezihack/colorlog"
//)
//
//var terraformProviderRepoUrl string
//
//func init() {
//	terraformProvider.Flags().StringVarP(&terraformProviderRepoUrl, "url", "u", "", "provider's repository URL")
//	rootCmd.AddCommand(terraformProvider)
//}
//
//var terraformProvider = &cobra.Command{
//	Use:   "terraform-provider",
//	Short: "Generate selefra provider from terraform's provider",
//	Long:  ``,
//	Run: func(cmd *cobra.Command, args []string) {
//		generator, err := generate_selefra_terraform_provider.NewGenerateTerraformProviderFromTerraformProviderRepoUrl(terraformProviderRepoUrl)
//		if err != nil {
//			colorlog.Error("new generator error: %s", err.Error())
//			return
//		}
//		err = generator.Run()
//		if err != nil {
//			colorlog.Error("generate error: %s", err.Error())
//		} else {
//			colorlog.Info("Done")
//		}
//	},
//}
