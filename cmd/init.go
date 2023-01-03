package cmd

// TODO 2023-1-3 15:44:51
//import (
//	"github.com/selefra/selefra-terraform-provider-scaffolding/generate_selefra_terraform_provider"
//	"github.com/spf13/cobra"
//	"log"
//)
//
//func init() {
//	rootCmd.AddCommand(initSelefraTerraformProvider)
//}
//
//var initSelefraTerraformProvider = &cobra.Command{
//	Use:   "init",
//	Short: "init selefra terraform provider project template",
//	Long:  ``,
//	Run: func(cmd *cobra.Command, args []string) {
//		generator, err := generate_selefra_terraform_provider.NewProviderGenerator(configFilePath)
//		if err != nil {
//			log.Fatal(err)
//			return
//		}
//		err = generator.Run()
//		if err != nil {
//			log.Fatal(err)
//		} else {
//			log.Println("Done")
//		}
//	},
//}
