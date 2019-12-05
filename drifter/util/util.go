package util

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/hashicorp/go-getter"
	"github.com/mattn/go-zglob"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reRepoName = regexp.MustCompile(`.*gruntwork-io\/([\w,\-,\_]+).*`)
var reVersion = regexp.MustCompile(`.*\?ref=(.*)$`)
var reTfFile = regexp.MustCompile(`\/[\w,\-,\_]+\.tf$`)

type ModuleSourceConfig struct {
	CurrentRelease string
	Owner          string
	Repository     string
	LatestRelease  string
	NeedsUpdate    bool
}

func ParseRepository(moduleSource string) (*ModuleSourceConfig, error) {
	conf := NewModuleSourceConfig()
	if strings.Contains(moduleSource, "github.com") && strings.Contains(moduleSource, "gruntwork-io") {
		repo, version := parseSource(moduleSource)
		conf.Repository = repo
		conf.CurrentRelease = version
		err := checkModuleRelease(conf)
		if err != nil {
			return nil, err
		}

	} else {
		conf.NeedsUpdate = false
	}
	return conf, nil
}

func checkModuleRelease(module *ModuleSourceConfig) error {
	client, ctx := createGitClientAndContext()
	release, _, e := client.Repositories.GetLatestRelease(ctx, module.Owner, module.Repository)

	if e != nil {
		return e
	}

	module.LatestRelease = release.GetTagName()

	if module.LatestRelease != module.CurrentRelease {
		module.NeedsUpdate = true
	}

	return nil
}

func CheckoutRepoToTempFolder(gitUrl string) (string, error) {
	dir, err := ioutil.TempDir("", "")
	actualDir := fmt.Sprintf("%s/drifter", dir)
	log.Printf("Checking out %s to %s", gitUrl, actualDir)
	if err != nil {
		return "", err
	}

	err = getter.GetAny(actualDir, gitUrl)
	if err != nil {
		return "", err
	}
	return actualDir, nil
}

func createGitClientAndContext() (*github.Client, context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_OAUTH_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, ctx
}

func NewModuleSourceConfig() *ModuleSourceConfig {
	return &ModuleSourceConfig{
		CurrentRelease: "",
		Owner:          "gruntwork-io",
		Repository:     "",
		LatestRelease:  "",
		NeedsUpdate:    false,
	}
}

func ListDirectories(rootDir string, recursive bool) []string {
	// Skipping the errs, I know :)
	abs, _ := filepath.Abs(rootDir)

	trimmed := strings.TrimRight(abs, "/")
	if recursive {
		rv := []string{}
		m := map[string]string{}
		globPattern := fmt.Sprintf("%s/**/*.tf", abs)
		files, _ := zglob.Glob(globPattern)
		for _, f := range files {
			if !strings.Contains(f, ".terragrunt-cache") {
				noTf := reTfFile.ReplaceAllString(f, "")
				if _, ok := m[noTf]; !ok {
					m[noTf] = noTf
					rv = append(rv, noTf)
				}
			}
		}
		return rv
	} else {
		return []string{trimmed}
	}
}

func parseSource(source string) (string, string) {
	return string(reRepoName.FindSubmatch([]byte(source))[1]), string(reVersion.FindSubmatch([]byte(source))[1])
}
