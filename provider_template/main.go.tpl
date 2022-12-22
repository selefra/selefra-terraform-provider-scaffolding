package main

import (
	"github.com/selefra/selefra-provider-sdk/grpc/serve"
	"{{.ModuleName}}/provider"
)

func main() {
	myProvider := provider.GetProvider()
	serve.Serve(myProvider.Name, myProvider)
}
