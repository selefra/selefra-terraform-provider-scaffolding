package generate_selefra_terraform_provider

import (
	"bytes"
	"context"
	"github.com/selefra/selefra-terraform-provider-scaffolding/provider_template/provider_template_v2_generate"
	"github.com/yezihack/colorlog"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type SchemaGenerator struct {
	config                      *Config
	selefraProviderRenderParams *SelefraProviderRenderParams
}

func NewSchemaGeneratorV2(config *Config, selefraProviderRenderParams *SelefraProviderRenderParams) *SchemaGenerator {
	return &SchemaGenerator{
		config:                      config,
		selefraProviderRenderParams: selefraProviderRenderParams,
	}
}

func (x *SchemaGenerator) Run(ctx context.Context) error {

	// Use the number of resources under resources.go to determine whether the corresponding schema is generated
	resources, err := x.GetResources()
	colorlog.Info("resources count: %d", len(resources))
	if err != nil {
		return err
	}
	newTableSlice := make([]*SelefraTableSchemaRenderParams, 0)
	for _, table := range x.selefraProviderRenderParams.TableSlice {
		if _, exists := resources[table.TableName]; exists {
			newTableSlice = append(newTableSlice, table)
		}
	}
	x.selefraProviderRenderParams.TableSlice = newTableSlice

	t, err := template.New("schema.go").Parse(string(provider_template_v2_generate.SelefraSchemaTemplate))
	if err != nil {
		colorlog.Error("parse schema.go template error: %s", err.Error())
		return err
	}

	buffer := bytes.Buffer{}
	if err = t.ExecuteTemplate(&buffer, "schema.go", x.selefraProviderRenderParams); err != nil {
		colorlog.Error("render schema.go error: %s", err.Error())
		return err
	}

	schemaGoOutputDirectory := filepath.Join(x.config.Output.Directory, "resources")
	_ = os.MkdirAll(schemaGoOutputDirectory, os.ModePerm)
	schemaGoOutputPath := filepath.Join(schemaGoOutputDirectory, "selefra_schema.go")
	if err := os.WriteFile(schemaGoOutputPath, buffer.Bytes(), os.ModePerm); err != nil {
		colorlog.Error("write file %s error: %s", schemaGoOutputPath, err.Error())
		return err
	}
	colorlog.Info("write file %s success", schemaGoOutputPath)
	return nil
}

func (x *SchemaGenerator) GetResources() (map[string]struct{}, error) {
	resourcesOutputDirectory := filepath.Join(x.config.Output.Directory, "provider")
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, resourcesOutputDirectory, func(info fs.FileInfo) bool {
		return true
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	resourceNameSet := make(map[string]struct{}, 0)
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				astutil.Apply(decl, func(cursor *astutil.Cursor) bool {
					funcDecl, ok := decl.(*ast.FuncDecl)
					if !ok {
						return true
					}
					tableName := x.parseTableName(funcDecl)
					if tableName != "" {
						resourceNameSet[tableName] = struct{}{}
					}
					return true
				}, nil)
			}

		}
	}
	return resourceNameSet, nil
}

func (x *SchemaGenerator) parseTableName(funcDecl *ast.FuncDecl) string {
	defer func() {
		recover()
	}()

	// TODO Improve accuracy and avoid misselection
	s := funcDecl.Body.List[0].(*ast.ReturnStmt).Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts[0].(*ast.KeyValueExpr).Value.(*ast.BasicLit).Value
	return strings.Trim(s, "\"")
}
