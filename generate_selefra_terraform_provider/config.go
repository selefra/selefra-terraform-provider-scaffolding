package generate_selefra_terraform_provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-git/go-git/v5"
	"github.com/selefra/selefra-provider-sdk/terraform/provider"
	"github.com/spf13/viper"
	"github.com/yezihack/colorlog"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config Some configuration information is required when the scaffold is running
type Config struct {

	// Selefra related configuration information, such as the name of the generated module, etc
	Selefra Selefra `mapstructure:"selefra" json:"selefra"`

	// Terraform-related parameter Settings, such as the Provider from which to generate the Selefra
	Terraform Terraform `mapstructure:"terraform" json:"terraform"`

	// Terraform-related parameter Settings, such as the Provider from which to generate the Selefra
	Output Output `mapstructure:"output" json:"output"`
}

// A copy of the configuration file is cached locally after each initialization, so that the next time you run generate,
// you can use the cache of the configuration file instead of generating a new copy, because generating a configuration
// file is a time-consuming operation, and adding this cache will greatly increase the speed of your application
var configJsonLocalPath = ".selefra_terraform_scaffolding_config.json"

// NewConfigFromLocalJson Read the configuration .file previously cached locally
func NewConfigFromLocalJson() (*Config, error) {
	file, err := os.ReadFile(configJsonLocalPath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	// The local cache is considered directly available and does not need to be checked
	return config, nil
}

// NewConfigFromEnv Try to generate a configuration file based on the parameters passed by the environment variable
func NewConfigFromEnv() (config *Config, err error) {

	// You can use environment variables to pass the path to a configuration file
	configPath := os.Getenv("SELEFRA_TERRAFORM_SCAFFOLDING_CONFIG_PATH")

	// Or just pass a repository URL for the Provider of Terraform
	// Both variables are acceptable here for version compatibility
	terraformProviderUrl := os.Getenv("TERRAFORM_PROVIDER_URL")
	if terraformProviderUrl == "" {
		terraformProviderUrl = os.Getenv("TERRAFORM_PROVIDER")
	}

	// The path of the configuration file has a higher priority. If the path of the configuration file is passed,
	// the path of the configuration file is preferentially used
	if configPath != "" {
		config, err = NewConfigFromPath(configPath)
		if err != nil {
			colorlog.Error("create config from path %s failed: %s", configPath, err.Error())
		} else {
			if err := checkConfig(config); err != nil {
				colorlog.Error("check config error: %s", err.Error())
				return nil, err
			}

			colorlog.Info("create config from path %s success", configPath)
			config.saveConfigToLocalJson()
			return
		}
	} else if terraformProviderUrl != "" {
		config, err = NewConfigFromTerraformProviderRepoUrl(terraformProviderUrl)
		if err != nil {
			colorlog.Error("create config from terraform provider url %s failed: %s", terraformProviderUrl, err.Error())
		} else {

			if err := checkConfig(config); err != nil {
				colorlog.Error("check config error: %s", err.Error())
				return nil, err
			}

			colorlog.Info("create config from terraform provider url %s success", terraformProviderUrl)
			config.saveConfigToLocalJson()
			return
		}
	}
	colorlog.Error("can not create config, please specify env variable TERRAFORM_PROVIDER_URL or SELEFRA_TERRAFORM_SCAFFOLDING_CONFIG_PATH")
	err = errors.New("config create failed")
	return
}

// NewConfigFromPath Creates a profile based on the specified profile path
func NewConfigFromPath(configFilePath string) (*Config, error) {
	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		colorlog.Error("read config file error: %s", err.Error())
		return nil, err
	}

	viperConfig := viper.New()
	viperConfig.SetConfigType("yaml")
	err = viperConfig.ReadConfig(bytes.NewReader(configBytes))
	if err != nil {
		colorlog.Error("viper read config file error: %s, content = %s", err.Error(), string(configBytes))
		return nil, err
	}

	config := new(Config)
	err = viperConfig.Unmarshal(&config)
	if err != nil {
		colorlog.Error("unmarshal config file error: %s, config file content = %s", err.Error(), string(configBytes))
		return nil, err
	}

	if err := checkConfig(config); err != nil {
		colorlog.Error("check config error: %s", err.Error())
		return nil, err
	}

	return config, nil
}

func NewConfigFromTerraformProviderRepoUrl(terraformProviderRepoUrl string) (*Config, error) {
	config := new(Config)
	config.Terraform.TerraformProvider.RepoUrl = terraformProviderRepoUrl
	if err := checkConfig(config); err != nil {
		colorlog.Error("check config error: %s", err.Error())
		return nil, err
	}

	return config, nil
}

// Do some checking and automatic configuration through this method
func checkConfig(config *Config) error {

	// Check that the URL of the Terraform Provider is set
	if config.Terraform.TerraformProvider.RepoUrl == "" {
		errorMsg := `The Terraform Provider URL is empty. You can solve the problem in the following centralized manner:
- If you specify the configuration file, you can specify the repository address of the Terraform Provider to be accessed in the terraform.provider.repo-url of the configuration file
- Or you can specify the environment variable TERRAFORM_PROVIDER_URL or TERRAFORM_PROVIDER`
		colorlog.Error(errorMsg)
		return ErrCheckConfigFailed
	}

	// If the module name is not configured, it is automatically generated. If it cannot be generated, an error message is displayed
	if config.getOrAutoDetectModuleName() == "" {
		errorMsg := `The module name cannot be read. Rectify the fault in one of the following ways:
- Make sure your repository is hosted on Github and synchronized locally using git clone
- Specify the module name in go.mod
- Use the environment variable SELEFRA_MODULE_NAME`
		colorlog.Error(errorMsg)
		return ErrCheckConfigFailed
	}

	// If the output path is not configured, a default is generated for it
	if config.Output.getDirectoryOrDefault() == "" {
		colorlog.Error("Use the environment variable SELEFRA_TERRAFORM_OUTPUT_DIRECTORY to specify the result output directory")
		return ErrCheckConfigFailed
	}
	colorlog.Info("workspace directory = %s", config.Output.getDirectoryOrDefault())

	// Try to resolve the name of the Provider that uses the Terraform
	if config.Terraform.TerraformProvider.GetOrParseProviderName() == "" {
		colorlog.Error("The Provider name cannot be resolved from the given Terraform Provider URL: %s", config.Terraform.TerraformProvider.RepoUrl)
		return ErrCheckConfigFailed
	}

	// It is the official provider
	if b, _ := config.Terraform.TerraformProvider.IsTerraformOfficialProvider(); b {
		// In the case of the official repository, the information for the downloadable file is generated from the official Registry
		files, err := config.Terraform.TerraformProvider.GetTerraformOfficialProviderFiles()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			colorlog.Error("You have specified an official provider, but I cannot automatically parse the corresponding provider file. Please make the provider file manually")
			return ErrCheckConfigFailed
		}
	}

	// It's a provider on github
	if b, _ := config.Terraform.TerraformProvider.IsGithubRepo(); len(config.Terraform.TerraformProvider.ExecuteFiles) == 0 && b {
		files, err := config.Terraform.TerraformProvider.RequestGithubReleaseFiles()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			colorlog.Error("You specified a provider hosted on Github, but I cannot automatically parse the corresponding provider file. Please specify the provider file manually")
			return ErrCheckConfigFailed
		}
	}

	if len(config.Terraform.TerraformProvider.ExecuteFiles) == 0 {
		colorlog.Error("it's a provider on github")
		return ErrCheckConfigFailed
	}

	colorlog.Info("Check whether the configuration information is correct")

	return nil
}

func (x *Config) saveConfigToLocalJson() {
	marshal, err := json.Marshal(x)
	if err != nil {
		colorlog.Error("marshal json error: %s", err.Error())
	}
	if err := os.WriteFile(configJsonLocalPath, marshal, os.ModePerm); err != nil {
		colorlog.Error("save config json error: %s", err.Error())
		return
	}
}

func (x *Config) IsResourceNeedGenerate(resourceName string) bool {
	if len(x.Terraform.TerraformProvider.Resources) == 0 {
		return true
	}
	// TODO 2022-12-14 17:00:44 convert to set
	for _, allowResourceName := range x.Terraform.TerraformProvider.Resources {
		if allowResourceName == resourceName {
			return true
		}
	}
	return false
}

type Selefra struct {
	ModuleName string `mapstructure:"module-name" json:"module_name"`
}

// If a module name is configured, use the given module name, otherwise try to detect environment information to automatically generate a module name for it
func (x *Config) getOrAutoDetectModuleName() string {
	if x.Selefra.ModuleName != "" {
		return x.Selefra.ModuleName
	}

	// You can set this parameter through environment variables
	x.Selefra.ModuleName = strings.TrimSpace(os.Getenv("SELEFRA_MODULE_NAME"))
	if x.Selefra.ModuleName != "" {
		return x.Selefra.ModuleName
	}

	// First try reading the module name from go.mod, which will only be used if the user changes the default module name
	x.Selefra.ModuleName = x.tryFindGitModuleNameFromGoMod()
	if x.Selefra.ModuleName == "" {
		// If the module name is not read from go.mod, the user has not changed it manually, and the repository URL is used as the module name
		x.Selefra.ModuleName = x.tryFindGitModuleNameFromLocalGitRepo()
	}

	return x.Selefra.ModuleName
}

// which will only be used if the user changes the default module name
func (x *Config) tryFindGitModuleNameFromGoMod() string {
	// Look up two levels for the go.mod file, assuming you're in the $root/bin directory
	fileBytes, err := os.ReadFile("go.mod")
	if err != nil {
		fileBytes, err = os.ReadFile("../go.mod")
	}
	if err != nil {
		return ""
	}
	// go.mod was read and tried to resolve the module name
	split := strings.Split(string(fileBytes), "\n")
	if len(split) < 1 {
		return ""
	}
	// example:
	// module github.com/selefra/selefra-terraform-provider-scaffolding
	split = strings.Split(strings.TrimSpace(split[0]), " ")
	if len(split) != 2 {
		return ""
	}
	if strings.ToLower(split[0]) != "module" {
		return ""
	}

	// black list for default module name, The modification takes effect only after you manually modify it
	if strings.TrimSpace(strings.ToLower(split[1])) == "github.com/selefra/selefra-provider-template" {
		return ""
	}

	// Obtaining the module name succeeded. Procedure
	colorlog.Info("read the module name %s from go.mod", split[1])
	return split[1]
}

// Try reading the Git repository and get the URL of the repository it is bound to to generate the module name
func (x *Config) tryFindGitModuleNameFromLocalGitRepo() string {
	// First try reading Git repository information from the current directory
	gitRepoPath := filepath.Join(x.Output.getDirectoryOrDefault(), ".git")
	open, err := git.PlainOpen(gitRepoPath)
	if err != nil {
		// If not, it attempts to read the repository information from the previous directory
		//colorlog.Error("try open .git repo %s error: %s", gitRepoPath, err.Error())
		gitRepoPath = filepath.Join(x.Output.getDirectoryOrDefault(), "../.git")
		open, err = git.PlainOpen(gitRepoPath)
	}
	if err != nil {
		colorlog.Error("Try open .git repo error: %s, module names for Selefra cannot be generated from git repositories", err.Error())
		return ""
	}
	colorlog.Info("Open git repo success: %s, parsing repo information...", gitRepoPath)
	remotes, err := open.Remotes()
	if err != nil {
		colorlog.Error("The remote url cannot be resolved from the repository, error: %s, module names for Selefra cannot be generated from git repositories", err.Error())
		return ""
	}
	colorlog.Info("The remote url for the git repository was read successfully and the module name that generated the selefra is being resolved...", gitRepoPath)
	for _, remote := range remotes {
		for _, gitRemoteUrl := range remote.Config().URLs {
			moduleName := convertGitUrl(gitRemoteUrl)
			if x.isOkGitRepoUrl(moduleName) {
				colorlog.Info("The Selefra module name %s is generated from the remote url %s of the Git repository", moduleName, gitRemoteUrl)
				return moduleName
			} else {
				colorlog.Info("The module name for Selefra cannot be generated from the Remote Url %s of the Git repository", gitRemoteUrl)
			}
		}
	}
	colorlog.Error("None of the Remote urls in the Git repository could generate the Selefra module name, and the automatic generation failed")
	return ""
}

// Determine if it is a legitimate Git repository
func (x *Config) isOkGitRepoUrl(repoUrl string) bool {
	if !strings.HasPrefix(strings.ToLower(repoUrl), "github.com/") {
		return false
	}
	split := strings.Split(repoUrl, "/")
	if len(split) != 3 {
		return false
	}
	return true
}

// Unify the remote addresses configured by http or git methods
func convertGitUrl(remoteUrl string) string {
	lowerRemoteUrl := strings.ToLower(remoteUrl)
	// is git protocol
	// example: git@github.com:selefra/selefra-terraform-provider-scaffolding.git
	if strings.HasPrefix(lowerRemoteUrl, "git@") {
		s := strings.ReplaceAll(remoteUrl, "git@", "")
		s = strings.ReplaceAll(s, ".git", "")
		s = strings.ReplaceAll(s, ":", "/")
		return s
	} else if strings.HasPrefix(lowerRemoteUrl, "https") || strings.HasPrefix(lowerRemoteUrl, "http") {
		// is http protocol
		// example: https://github.com/selefra/selefra-terraform-provider-scaffolding.git
		s := strings.ReplaceAll(remoteUrl, ".git", "")
		s = strings.ReplaceAll(s, "http://", "")
		s = strings.ReplaceAll(s, "https://", "")
		return s
	} else {
		return remoteUrl
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type Terraform struct {
	TerraformProvider TerraformProvider `mapstructure:"provider" json:"terraform_provider"`
}

// ------------------------------------------------- --------------------------------------------------------------------

// TerraformProvider Set the based Terraform provider's parameters
type TerraformProvider struct {

	// provider's warehouse
	RepoUrl string `mapstructure:"repo-url" json:"repo_url"`

	// This parameter is required when the provider starts
	Config string `mapstructure:"config" json:"config"`

	// Provider executable file
	ExecuteFiles []*provider.TerraformProviderFile `mapstructure:"execute-files" json:"execute_files"`

	// Resources to be generated. If not set, all resources are generated by default
	Resources []string `mapstructure:"resources" json:"resources"`

	providerName string `json:"provider_name"`
}

// IsGithubRepo Determines whether the specified repository is a GitHub repository
func (x *TerraformProvider) IsGithubRepo() (bool, error) {
	parse, err := url.Parse(x.RepoUrl)
	if err != nil {
		return false, err
	}
	return strings.ToLower(parse.Hostname()) == "github.com", nil
}

// RequestGithubReleaseFiles A list of the latest releases from GitHub's repository
func (x *TerraformProvider) RequestGithubReleaseFiles() ([]*provider.TerraformProviderFile, error) {
	// use cache
	if len(x.ExecuteFiles) != 0 {
		return x.ExecuteFiles, nil
	}
	parse, err := url.Parse(x.RepoUrl)
	if err != nil {
		return nil, err
	}
	// Use the GitHub API to request the latest Release of the repository
	targetUrl := "https://api.github.com/repos" + parse.Path + "/releases/latest"
	response := request(targetUrl)
	if response == nil {
		return nil, fmt.Errorf("request github repo latest releases failed")
	}
	r := &GithubLatestReleasesResponse{}
	err = json.Unmarshal(response.Body(), &r)
	if err != nil {
		return nil, fmt.Errorf("github repo latest releases response json unmarshal failed: %s", err.Error())
	}
	// make cache
	x.ExecuteFiles = r.ParseProviderFileSlice()

	colorlog.Info("request github release files %s success, find %d releases files", targetUrl, len(x.ExecuteFiles))

	return x.ExecuteFiles, nil
}

// IsTerraformOfficialProvider Check whether the current provider is an official provider
func (x *TerraformProvider) IsTerraformOfficialProvider() (bool, error) {
	parse, err := url.Parse(x.RepoUrl)
	if err != nil {
		return false, err
	}
	if strings.ToLower(parse.Hostname()) != "github.com" {
		return false, nil
	}
	return strings.HasPrefix(parse.Path, "/hashicorp/"), nil
}

// GetTerraformOfficialProviderFiles Get the official provider executable file list
func (x *TerraformProvider) GetTerraformOfficialProviderFiles() ([]*provider.TerraformProviderFile, error) {

	// use cache
	if len(x.ExecuteFiles) != 0 {
		colorlog.Info("The Provider Release file is specified in the configuration file, which does not need to be automatically parsed and generated")
		return x.ExecuteFiles, nil
	}

	// 1. Obtain the latest version of the provider
	providerName := x.GetOrParseProviderName()
	if providerName == "" {
		colorlog.Error("The Provider name cannot be resolved from the given Terraform Provider URL: %s", x.RepoUrl)
		return nil, ErrCheckConfigFailed
	}
	targetUrl := "https://releases.hashicorp.com/" + providerName
	response := request(targetUrl)
	if response == nil {
		colorlog.Error("An attempt to automatically generate Release information from the official Terraform Provider %s failed. Please try again")
		return nil, ErrCheckConfigFailed
	}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response.Body()))
	if err != nil {
		colorlog.Error("goquery failed to parse html: %s, response = %s", err.Error(), response.String())
		return nil, ErrCheckConfigFailed
	}
	latestVersionFilePage := ""
	document.Find("li>a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if latestVersionFilePage == "" && exists && !strings.HasPrefix(href, "../") {
			latestVersionFilePage = "https://releases.hashicorp.com" + href
		}
	})
	colorlog.Info("Terraform provider %s, find latest version %s", x.GetOrParseProviderName(), latestVersionFilePage)

	// Random hibernation to avoid too frequent requests to Terraform's official repository
	time.Sleep(time.Second * time.Duration(rand.Intn(3)+3))

	// 2. Obtain the file of the latest version
	response = request(latestVersionFilePage)
	if response == nil {
		colorlog.Error("Failed to obtain the release file of Terraform Provider %s's version %s", providerName, latestVersionFilePage)
		return nil, ErrCheckConfigFailed
	}
	document, err = goquery.NewDocumentFromReader(bytes.NewReader(response.Body()))
	if err != nil {
		colorlog.Error("goquery fails to parse HTMl, error message: %s, response = %s", err.Error(), response.String())
		return nil, ErrCheckConfigFailed
	}
	providerFileSlice := make([]*provider.TerraformProviderFile, 0)
	document.Find("a[data-product]").Each(func(i int, selection *goquery.Selection) {
		version, _ := selection.Attr("data-version")
		os, _ := selection.Attr("data-os")
		arch, _ := selection.Attr("data-arch")
		downloadUrl, _ := selection.Attr("href")
		providerFileSlice = append(providerFileSlice, &provider.TerraformProviderFile{
			ProviderVersion: version,
			DownloadUrl:     downloadUrl,
			Arch:            arch,
			OS:              os,
		})
	})

	// make cache
	x.ExecuteFiles = providerFileSlice

	colorlog.Info("request Terraform provider %s releases success, find %d releases files", x.GetOrParseProviderName(), len(x.ExecuteFiles))
	for _, file := range x.ExecuteFiles {
		colorlog.Info("\t\tArch: %s", file.Arch)
		colorlog.Info("\t\tOS: %s", file.OS)
		colorlog.Info("\t\tDownload URL: %s", file.DownloadUrl)
	}

	if len(x.ExecuteFiles) == 0 {
		colorlog.Error("The version information of Terraform Provider is not resolved, URL = %s, response = %s", latestVersionFilePage, response.String())
	}

	return x.ExecuteFiles, nil
}

// GetOrParseProviderName Resolve the provider name from the repository url
func (x *TerraformProvider) GetOrParseProviderName() string {
	if x.providerName != "" {
		return x.providerName
	}
	split := strings.Split(strings.Trim(x.RepoUrl, "/"), "/")
	if len(split) < 1 {
		return ""
	}
	x.providerName = split[len(split)-1]
	return x.providerName
}

func (x *TerraformProvider) ParseProviderShortName() string {
	return strings.ReplaceAll(x.GetOrParseProviderName(), "terraform-provider-", "")
}

// ------------------------------------------------- --------------------------------------------------------------------

// Output Set output parameters
type Output struct {

	// The directory to which the generated results are output
	Directory string `mapstructure:"directory" json:"directory"`
}

// If the output directory is configured, the user configured one is used, otherwise a default is generated for it
func (x *Output) getDirectoryOrDefault() string {
	if x.Directory != "" {
		return x.Directory
	}

	x.Directory = os.Getenv("SELEFRA_TERRAFORM_OUTPUT_DIRECTORY")
	if x.Directory == "" {
		x.Directory = "./"
	}

	return x.Directory
}

// ------------------------------------------------- --------------------------------------------------------------------

type GithubLatestReleasesResponse struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

// ParseProviderFileSlice The corresponding download file is parsed from the GitHub API response
func (x *GithubLatestReleasesResponse) ParseProviderFileSlice() []*provider.TerraformProviderFile {
	providerFileSlice := make([]*provider.TerraformProviderFile, 0)
	osSet, archSet := getGoAllowArchAndOS()
	for _, assert := range x.Assets {
		os, arch := tryFindOsAndArch(assert.Name, osSet, archSet)
		if os == "" || arch == "" {
			continue
		}
		providerFileSlice = append(providerFileSlice, &provider.TerraformProviderFile{
			ProviderVersion: x.Name,
			DownloadUrl:     assert.BrowserDownloadURL,
			OS:              os,
			Arch:            arch,
		})
	}
	return providerFileSlice
}

// Try to identify the operating system and arch from the file name, if it contains
func tryFindOsAndArch(name string, osSet, archSet map[string]struct{}) (os string, arch string) {
	wordRightIndex := -1
	inWord := false
	for index := len(name) - 1; index >= 0; index-- {
		c := name[index]
		isWordChar := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if isWordChar && !inWord {
			// find word end
			inWord = true
			wordRightIndex = index + 1
		} else if !isWordChar && inWord {
			// find word begin
			inWord = false
			word := name[index+1 : wordRightIndex]
			if _, exists := osSet[word]; exists && os == "" {
				os = word
			} else if _, exists := archSet[word]; exists && arch == "" {
				arch = word
			}
		}
	}
	// ignore the left border, think no possible
	return
}

// All operating systems and platforms supported by Go
func getGoAllowArchAndOS() (osSet map[string]struct{}, archSet map[string]struct{}) {
	s := `aix/ppc64
android/386    
android/amd64  
android/arm    
android/arm64  
darwin/amd64   
darwin/arm64   
dragonfly/amd64
freebsd/386    
freebsd/amd64  
freebsd/arm    
freebsd/arm64  
illumos/amd64  
ios/amd64      
ios/arm64      
js/wasm        
linux/386      
linux/amd64    
linux/arm      
linux/arm64
linux/loong64
linux/mips
linux/mips64
linux/mips64le
linux/mipsle
linux/ppc64
linux/ppc64le
linux/riscv64
linux/s390x
netbsd/386
netbsd/amd64
netbsd/arm
netbsd/arm64
openbsd/386
openbsd/amd64
openbsd/arm
openbsd/arm64
openbsd/mips64
plan9/386
plan9/amd64
plan9/arm
solaris/amd64
windows/386
windows/amd64
windows/arm
windows/arm64
`
	osSet = make(map[string]struct{}, 0)
	archSet = make(map[string]struct{}, 0)
	for _, item := range strings.Split(s, "\n") {
		split := strings.Split(item, "/")
		if len(split) != 2 {
			continue
		}
		os := strings.TrimSpace(split[0])
		arch := strings.TrimSpace(split[1])
		osSet[os] = struct{}{}
		archSet[arch] = struct{}{}
	}
	return osSet, archSet
}

// ------------------------------------------------- --------------------------------------------------------------------
