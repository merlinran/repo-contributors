package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

type Filter func(*github.User) bool

func and(a Filter, b Filter) Filter {
	return func(u *github.User) bool {
		return a(u) && b(u)
	}
}

func or(a Filter, b Filter) Filter {
	return func(u *github.User) bool {
		return a(u) || b(u)
	}
}

type RepoContributersFetcher struct {
	client *github.Client
	filter Filter
}

func (fetcher *RepoContributersFetcher) processRepo(r string) {
	fmt.Printf("Start processing '%v'\n", r)
	a := strings.Split(r, "/")
	if len(a) != 2 {
		fmt.Printf("'%v' is incorrect\n", r)
		return
	}
	owner, repo := a[0], a[1]
	ch := fetcher.fetch(owner, repo)
	f, err := os.OpenFile(repo+".csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	defer f.Close()
	for ret := range ch {
		f.Write([]byte("\""))
		_, err := f.Write([]byte(strings.Join(ret, "\",\"")))
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		f.Write([]byte("\"\n"))
	}
}

func v(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

func (fetcher *RepoContributersFetcher) fetch(owner, repo string) <-chan []string {
	ch := make(chan []string)
	stats, _, err := fetcher.client.Repositories.ListContributorsStats(owner, repo)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	go func() {
		ch <- fetcher.getHeader()
		for _, s := range stats {
			c := s.Author
			u, _, err := fetcher.client.Users.Get(*c.Login)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			if fetcher.filter(u) {
				ch <- fetcher.getUser(u, *s.Total)
			}
		}
		close(ch)
	}()

	return ch
}

func (fetcher *RepoContributersFetcher) getHeader() []string {
	return []string{"Name", "Email", "Location", "URL", "Points"}
}

func (fetcher *RepoContributersFetcher) getUser(u *github.User, points int) []string {
	return []string{v(u.Name), v(u.Email),
		v(u.Location), v(u.HTMLURL), strconv.Itoa(points)}
}
