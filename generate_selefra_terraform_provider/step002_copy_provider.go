package generate_selefra_terraform_provider

import (
	"bytes"
	"github.com/yezihack/colorlog"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CopyProvider struct {
	config *Config
}

func NewCopyProvider(config *Config) *CopyProvider {
	return &CopyProvider{
		config: config,
	}
}

func (x *CopyProvider) Run() error {

	sourceDirectory := filepath.Join(x.config.Output.Directory, "provider")
	destinationDirectory := filepath.Join(x.config.Output.Directory, "resources")

	// delete
	err := os.RemoveAll(destinationDirectory)
	if err != nil {
		colorlog.Error("remove directory %s failed: %s", destinationDirectory, err.Error())
	} else {
		colorlog.Info("remove directory %s successfully", destinationDirectory)
	}

	err = filepath.Walk(sourceDirectory, func(sourcePath string, info os.FileInfo, err error) error {
		destinationPath := x.computeDestinationPath(sourceDirectory, destinationDirectory, sourcePath)
		if info.IsDir() {
			err := os.MkdirAll(destinationPath, os.ModeDir|os.ModePerm)
			if err != nil {
				colorlog.Error("create directory %s failed: %s", destinationPath, err.Error())
			} else {
				colorlog.Info("create directory %s success", destinationPath)
			}
			return nil
		}
		fileBytes, err := x.processGoFile(sourcePath)
		if err != nil {
			colorlog.Error("process file %s failed: %s", sourcePath, err.Error())
			return err
		}
		err = os.WriteFile(destinationPath, fileBytes, os.ModePerm)
		if err != nil {
			colorlog.Error("copy file %s failed: %s", sourcePath, err.Error())
			return err
		} else {
			colorlog.Info("copy file %s to %s success", sourcePath, destinationPath)
			return nil
		}
	})

	return err
}

func (x *CopyProvider) computeDestinationPath(sourceDirectory, destinationDirectory, sourcePath string) string {
	// TODO maybe have problem ?
	sourcePath = strings.ReplaceAll(sourcePath, "\\", "/")
	sourceDirectory = strings.ReplaceAll(sourceDirectory, "\\", "/")
	index := strings.Index(sourcePath, sourceDirectory)
	if index == -1 {
		colorlog.Error("destination directory error, sourceDirectory = %s, destinationDirectory = %s, sourcePath = %s", sourceDirectory, destinationDirectory, sourcePath)
		return ""
	}
	if index+len(sourceDirectory)+1 > len(sourcePath) {
		return destinationDirectory
	}
	return filepath.Join(destinationDirectory, sourcePath[index+len(sourceDirectory)+1:])
}

func (x *CopyProvider) processGoFile(filepath string) ([]byte, error) {
	if !strings.HasSuffix(filepath, ".go") {
		return os.ReadFile(filepath)
	}
	colorlog.Info("begin ast parse go file %s...", filepath)
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	f.Name.Name = "resources"
	//astutil.Apply(f, func(cursor *astutil.Cursor) bool {
	//	// TODO 2022-12-30 19:35:59
	//	packageNode, ok := cursor.Node().(*ast.Package)
	//	if ok {
	//		packageNode.Name = "resources"
	//	}
	//	return true
	//}, nil)

	buff := bytes.Buffer{}
	err = printer.Fprint(&buff, fileSet, f)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
