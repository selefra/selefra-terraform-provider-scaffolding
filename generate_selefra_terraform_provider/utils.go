package generate_selefra_terraform_provider

import (
	"os"
	"strings"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func escapeStringForQuote(s string) string {
	buff := strings.Builder{}
	for index, char := range s {
		if char == '"' && index > 0 && s[index-1] != '\\' {
			buff.WriteString("\\")
		}
		buff.WriteRune(char)
	}
	return buff.String()
}
