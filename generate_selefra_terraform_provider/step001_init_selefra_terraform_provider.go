package generate_selefra_terraform_provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_init"
	"github.com/yezihack/colorlog"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type SelefraTerraformProviderInit struct {
	config          *Config
	schemaIRManager *SchemaIRManager
}

func NewSelefraTerraformProviderInit(config *Config) *SelefraTerraformProviderInit {
	return &SelefraTerraformProviderInit{
		config:          config,
		schemaIRManager: NewSchemaIRManager(config),
	}
}

func (x *SelefraTerraformProviderInit) Run(ctx context.Context) error {

	// generate schema.json
	if err := x.schemaIRManager.GenerateIRAndSave(ctx); err != nil {
		return err
	}

	// rewrite provider.go
	if err := x.RewirteProviderGo(); err != nil {
		return err
	}

	// rewrite resource.go
	if err := x.RewriteResourcesGo(); err != nil {
		return err
	}

	if err := x.RewriteGoMod(); err != nil {
		return err
	}

	return nil
}

func (x *SelefraTerraformProviderInit) RewriteGoMod() error {
	goModPath := filepath.Join(x.config.Output.Directory, "go.mod")
	file, err := os.ReadFile(goModPath)
	if err != nil {
		colorlog.Error("can not open file %s, error msg: %s", goModPath, err.Error())
		return err
	}
	newGoModFile := strings.ReplaceAll(string(file), "module github.com/selefra/selefra-provider-template", "module "+x.config.Selefra.ModuleName)
	err = os.WriteFile(goModPath, []byte(newGoModFile), os.ModePerm)
	if err != nil {
		colorlog.Error("rewrite go.mod file error: %s", err.Error())
	} else {
		colorlog.Info("rewrite go.mod file success")
	}
	return err
}

func (x *SelefraTerraformProviderInit) RewirteProviderGo() error {
	providerOutputDirectory := filepath.Join(x.config.Output.Directory, "provider")
	pathOutputPath := filepath.Join(providerOutputDirectory, "provider.go")
	if exists, err := PathExists(pathOutputPath); err == nil && exists {
		colorlog.Info("file %s already exists, so do not regenerate", pathOutputPath)
		return nil
	}

	t, err := template.New("provider.go").Parse(string(provider_template_v2_init.ProviderTemplate))
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	renderParams := &InitProviderGoRenderParams{
		SelefraProviderName:               x.config.Terraform.TerraformProvider.ParseProviderShortName(),
		ModuleName:                        x.config.Selefra.ModuleName,
		TerraformProviderExecuteFileSlice: x.config.Terraform.TerraformProvider.ExecuteFiles,
	}
	if err = t.ExecuteTemplate(&buffer, "provider.go", renderParams); err != nil {
		return err
	}
	_ = os.MkdirAll(providerOutputDirectory, os.ModePerm)
	if err := os.WriteFile(pathOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (x *SelefraTerraformProviderInit) RewriteResourcesGo() error {
	// Load the existing resource
	resourcesOutputDirectory := filepath.Join(x.config.Output.Directory, "provider")
	resourcesOutputPath := filepath.Join(resourcesOutputDirectory, "resources.go")

	existsResourceSet := x.ParseExistsResourceSet()
	colorlog.Info("load exists resource %d", len(existsResourceSet))

	terraformProviderSchemaIR, err := x.schemaIRManager.readTerraformSchemaIR()
	if err != nil {
		colorlog.Error("read terraform schema IR failed: %s", err.Error())
		return err
	}

	resourceNeedGenerateCount := 0
	alreadyExistsCount := 0
	newAddExistsCount := 0
	resourceCodeBuff := bytes.Buffer{}
	ignoredResourceNameSlice := make([]string, 0)

	// sort by dictionary order
	sort.Slice(terraformProviderSchemaIR.Resources, func(i, j int) bool {
		return terraformProviderSchemaIR.Resources[i].ResourceName < terraformProviderSchemaIR.Resources[j].ResourceName
	})

	for _, terraformResourceSchemaIR := range terraformProviderSchemaIR.Resources {
		if !x.config.IsResourceNeedGenerate(terraformResourceSchemaIR.ResourceName) {
			continue
		}
		resourceNeedGenerateCount++
		if _, exists := existsResourceSet[terraformResourceSchemaIR.ResourceName]; exists {
			alreadyExistsCount++
			//colorlog.Info("resource %s already exists, so ignored", terraformResourceSchemaIR.ResourceName)
			ignoredResourceNameSlice = append(ignoredResourceNameSlice, terraformResourceSchemaIR.ResourceName)
			continue
		}
		s := `// terraform resource: %s
func GetResource_%s() *selefra_terraform_schema.SelefraTerraformResource {
	return &selefra_terraform_schema.SelefraTerraformResource{
		SelefraTableName:      "%s",
		TerraformResourceName: "%s",
		Description:           "%s",
		SubTables:             nil,
		ListResourceParamsFunc: func(ctx context.Context, clientMeta *schema.ClientMeta, taskClient any, task *schema.DataSourcePullTask, resultChannel chan<- any) ([]*selefra_terraform_schema.ResourceRequestParam, *schema.Diagnostics) {
			// TODO
			return nil, nil
		},
	}
}

`
		resourceCodeString := fmt.Sprintf(s, terraformResourceSchemaIR.ResourceName, terraformResourceSchemaIR.ResourceName, terraformResourceSchemaIR.ResourceName, terraformResourceSchemaIR.ResourceName, terraformResourceSchemaIR.Description)
		resourceCodeBuff.WriteString(resourceCodeString)
		newAddExistsCount++
	}

	prefix := ""
	if exists, err := PathExists(resourcesOutputPath); !exists || err != nil {
		prefix = `package provider

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/terraform/selefra_terraform_schema"
)


`
	}

	// append code to resource.go
	file, err := os.OpenFile(resourcesOutputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		colorlog.Error("open %s error: %s", resourcesOutputPath, err.Error())
		return err
	}
	_, err = file.WriteString(prefix + resourceCodeBuff.String())
	if err != nil {
		colorlog.Error("write %s error: %s", resourcesOutputPath, err.Error())
		return err
	}

	if len(ignoredResourceNameSlice) != 0 {
		colorlog.Info("ignored resource: %s", ignoredResourceNameSlice)
	}

	colorlog.Info("init resource.go success: ")
	colorlog.Info("\t\tTotal Need Generate Resource Count: %d", resourceNeedGenerateCount)
	colorlog.Info("\t\tAlready Exists Resource Count: %d", alreadyExistsCount)
	colorlog.Info("\t\tNew Add Resource Count: %d", newAddExistsCount)
	return nil
}

func (x *SelefraTerraformProviderInit) ParseExistsResourceSet() map[string]struct{} {
	existsResourceSet := make(map[string]struct{})
	resourceGoOutputDirectory := filepath.Join(x.config.Output.Directory, "provider")
	resourceGoOutputPath := filepath.Join(resourceGoOutputDirectory, "resources.go")
	if exists, err := PathExists(resourceGoOutputPath); err != nil || !exists {
		return existsResourceSet
	}
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, resourceGoOutputPath, nil, parser.ParseComments)
	if err != nil {
		colorlog.Error("parse resource.go file error: %s", err.Error())
		return existsResourceSet
	}
	astutil.Apply(f, func(cursor *astutil.Cursor) bool {
		kvExpr, ok := cursor.Node().(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		keyIdent, ok := kvExpr.Key.(*ast.Ident)
		if !ok {
			return true
		}
		if keyIdent.Name != "TerraformResourceName" {
			return true
		}
		valueIdent, ok := kvExpr.Value.(*ast.BasicLit)
		if !ok {
			return true
		}
		// maybe wrong, if wrong, I back here change judge rule
		resourceName := strings.Trim(valueIdent.Value, "\"")
		existsResourceSet[resourceName] = struct{}{}
		return true
	}, nil)
	return existsResourceSet
}

// ------------------------------------------------- --------------------------------------------------------------------

type InitProviderGoRenderParams struct {
	SelefraProviderName               string
	ModuleName                        string
	TerraformProviderExecuteFileSlice []*provider.TerraformProviderFile
}

type InitProviderTablesGoRenderParams struct {
	TableGeneratorNameSlice []string
	ModuleName              string
}

// ------------------------------------------------- --------------------------------------------------------------------
