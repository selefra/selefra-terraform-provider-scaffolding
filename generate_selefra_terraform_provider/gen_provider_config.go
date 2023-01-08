package generate_selefra_terraform_provider

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/yezihack/colorlog"
	"math/rand"
	"time"
)

func GenerateProviderExecuteFiles(terraformReleasePageUrl string) {
	response := request(terraformReleasePageUrl)
	if response == nil {
		colorlog.Error("request terraform release page %s failed", terraformReleasePageUrl)
		return
	}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response.Body()))
	if err != nil {
		colorlog.Error("parse terraform release page response document error: %w", err)
		return
	}
	document.Find("a[data-product=terraform-provider-aws]").Each(func(i int, selection *goquery.Selection) {
		version, _ := selection.Attr("data-version")
		os, _ := selection.Attr("data-os")
		arch, _ := selection.Attr("data-arch")
		downloadUrl, _ := selection.Attr("href")
		s := `      - provider-version: "%s"
        download-url: "%s"
        sha256-sum: ""
        arch: "%s"
        os: "%s"`
		fmt.Println(fmt.Sprintf(s, version, downloadUrl, arch, os))
	})
}

// Request the given address, If the request fails, return Nil
func request(targetUrl string) *resty.Response {
	colorlog.Info("Start sending request %s", targetUrl)
	for tryTimes := 1; tryTimes <= 30; tryTimes++ {
		if tryTimes != 1 {
			colorlog.Info("Request URL %s, start retry...", targetUrl)
		}
		start := time.Now()
		response, err := resty.
			New().SetTimeout(time.Minute*3).
			R().
			SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9").
			//SetHeader("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36").
			SetHeader("user-agent", "HashiCorp Terraform/v1.3.6 (+https://www.terraform.io)").
			//SetHeader("referer", "https://releases.hashicorp.com/terraform-provider-azurerm").
			Get(targetUrl)
		if err != nil {
			colorlog.Error("try times = %d, request url %s error: %s", tryTimes, targetUrl, err.Error())
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(4000)+1000))
			continue
		}
		cost := time.Now().Sub(start)
		colorlog.Info("Succeeded in requesting %s. cost time: %s", targetUrl, cost.String())
		return response
	}
	return nil
}
