package generate_selefra_terraform_provider

import (
	"testing"
)

func Test_escapeStringForQuote(t *testing.T) {
	s := "asdasdasdasdasdasd\""
	quote := escapeStringForQuote(s)
	t.Log(quote)
}
