// Code generated by https://github.com/selefra/selefra-terraform-provider-scaffolding DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***
package main

import (
	"github.com/selefra/selefra-provider-sdk/grpc/serve"
	"{{.ModuleName}}/resources"
)

func main() {

	myProvider := resources.GetSelefraProvider()
	serve.Serve(myProvider.Name, myProvider)

}
