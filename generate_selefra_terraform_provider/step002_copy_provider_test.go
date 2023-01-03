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
