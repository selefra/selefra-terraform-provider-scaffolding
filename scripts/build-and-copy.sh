#!/bin/bash

git pull
rm ../selefra-provider-test/bin/selefra-terraform-provider-scaffolding
go build
mv selefra-terraform-provider-scaffolding ../selefra-provider-test/bin/selefra-terraform-provider-scaffolding

