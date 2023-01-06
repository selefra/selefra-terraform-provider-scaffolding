package resources

import (
	"github.com/selefra/selefra-provider-sdk/test_helper"
	"os"
	"strings"
	"testing"
)

func Test_Provider(t *testing.T) {
	provider := GetSelefraProvider()
	split := strings.Split(os.Getenv("SELEFRA_TEST_TABLES"), ",")
	test_helper.RunProviderPullTables(provider, "log: info", "./", split...)
}
