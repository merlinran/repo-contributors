package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

var chineseSpeakingLocations = []string{
	"china",
	"beijing",
	"shanghai",
	"hangzhou",
	"chengdu",
	"shenzhen",
	"guangzhou",

	"taipei",
	"taiwan",
	"hong kong",
	"singapore",

	"everywhere",
	"earth",
}

func inChina(u *github.User) bool {
	loc := v(u.Location)
	if loc == "" {
		// don't want to miss anyone
		return true
	}
	for _, city := range chineseSpeakingLocations {
		if strings.Contains(strings.ToLower(loc), city) {
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
