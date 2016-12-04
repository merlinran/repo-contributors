package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitHubToken string   `yaml:"githubToken"`
	Repos       []string `yaml:",flow"`
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	client := ghClient(config.GitHubToken)
	fetcher := RepoContributersFetcher{client, inChina}
	for _, r := range config.Repos {
		fetcher.processRepo(r)
	}
}

var chinaCities = []string{
	"beijing",
	"shanghai",
	"hangzhou",
	"chengdu",
	"shenzhen",
	"guangzhou",

	"taiwan",
	"hong kong",
	"singapore",
}

var locRegexp = regexp.MustCompile("[^,]+(, [^,]+)*$")

func inChina(u *github.User) bool {
	loc := v(u.Location)
	if loc == "" {
		// don't want to miss anyone
		return true
	}
	matched := locRegexp.FindStringSubmatch(loc)
	if len(matched) == 2 && matched[1] == ", China" {
		return true
	}
	for _, city := range chinaCities {
		if strings.HasPrefix(strings.ToLower(loc), city) {
			return true
		}
	}
	fmt.Printf("Skipping user from '%v'\n", loc)
	return false
}

func ghClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}

func readConfig() (*Config, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, eread := ioutil.ReadAll(f)
	if eread != nil {
		return nil, eread
	}
	var config Config
	eload := yaml.Unmarshal(b, &config)
	if eload != nil {
		return nil, eload
	}
	return &config, nil
}
