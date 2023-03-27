package generate_selefra_terraform_provider

//import (
//	"bytes"
//	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
//	"github.com/yezihack/colorlog"
//	"os"
//	"path"
//	"text/template"
//)
//
//// GoModGenerator Used to render go.mod
//type GoModGenerator struct {
//	config *Config
//}
//
//func NewGoModGenerator(config *Config) *GoModGenerator {
//	return &GoModGenerator{
//		config: config,
//	}
//}
//
//func (x *GoModGenerator) Run() error {
//	return x.Render()
//}
//
//func (x *GoModGenerator) Render() error {
//	colorlog.Info("begin render go.mod...")
//	t, err := template.New("go.mod").Parse(string(provider_template_v2_generate.GoModTemplate))
//	if err != nil {
//		colorlog.Error("parse go.mod template error: %s", err.Error())
//		return err
//	}
//	buffer := bytes.Buffer{}
//	params := &GoModRenderParams{
//		ModuleName: x.config.Selefra.ModuleName,
//	}
//	if err = t.ExecuteTemplate(&buffer, "go.mod", params); err != nil {
//		colorlog.Error("render go.mod template error: %s", err.Error())
//		return err
//	}
//
//	goModOutputDirectory := filepath.Join(x.config.Output.Directory, "resources")
//	_ = os.MkdirAll(goModOutputDirectory, os.ModePerm)
//	goModOutputPath := filepath.Join(goModOutputDirectory, "go.mod")
//	if err := os.WriteFile(goModOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
//		colorlog.Error("write file %s error: %s", goModOutputPath, err.Error())
//		return err
//	}
//	colorlog.Info("render go.mod success, write to %s success", goModOutputPath)
//	return nil
//}
//
//type GoModRenderParams struct {
//	ModuleName string
//}
