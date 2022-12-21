package main

import (
	"github.com/selefra/selefra-provider-sdk/grpc/serve"
	"{{ProviderModuleUrl}}/provider"
)

func main() {
	myProvider := provider.GetProvider()
	serve.Serve(myProvider.Name, myProvider)
}
