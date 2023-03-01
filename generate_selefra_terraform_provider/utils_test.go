package generate_selefra_terraform_provider

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_escapeStringForQuote(t *testing.T) {
	s := "asdasdasdasdasdasd\""
	quote := escapeStringForQuote(s)
	t.Log(quote)
}

func Test_processDescription(t *testing.T) {

	// case 001. normal string
	s := "hello, world"
	excepted := "`hello, world`"
	answer := processDescription(s)
	assert.Equal(t, excepted, answer)

	// case 002.
	s = "hello', world\" \\"
	excepted = "`hello', world\" \\`"
	answer = processDescription(s)
	assert.Equal(t, excepted, answer)

	// case 003.
	s = "hello',`"
	excepted = "\"hello',`\""
	answer = processDescription(s)
	assert.Equal(t, excepted, answer)

	// case 004.
	s = `\n\n\n\n`
	excepted = "`\\n\\n\\n\\n`"
	answer = processDescription(s)
	assert.Equal(t, excepted, answer)

}

