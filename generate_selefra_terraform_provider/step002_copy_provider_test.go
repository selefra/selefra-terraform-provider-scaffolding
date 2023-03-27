package generate_selefra_terraform_provider

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyProvider_Run(t *testing.T) {
	err := NewCopyProvider(&Config{
		Output: Output{
			Directory: "./test/",
		},
	}).Run()
	assert.Nil(t, err)
}

func TestCopyProvider_computeDestinationPath(t *testing.T) {
	destPath := NewCopyProvider(&Config{
		Output: Output{
			Directory: "./test/",
		},
	}).computeDestinationPath("provider", "resources", "provider/provider.go")
	t.Log(destPath)
	assert.NotEqual(t, "", destPath)
}
