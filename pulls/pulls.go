package pulls

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  repos []*github.Repository, m map[string]int) (error) {
	var list []*github.PullRequest
	for _, repo := range repos {
		repoName := repo.GetName()
		repoOwner := repo.GetOwner().GetLogin()
		opt := &github.PullRequestListOptions{State: "all", Base: "master", ListOptions: github.ListOptions{PerPage: 30}}
		for {
			l, resp, err := client.PullRequests.List(ctx, repoOwner, repoName, opt)
			if err != nil {
				return err
			}
			list = append(list, l...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}
	GetPullsTimes(list, m, username)
	return nil
}

func GetPullsTimes(list []*github.PullRequest, m map[string]int, username string) {
	for _, pull := range list {
		if pull.GetUser().GetLogin() == username {
			time := pull.GetCreatedAt().Format("2006-01-02")
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
}

func PullsBase(m map[string]int) *charts.Bar {
	var keys []string
	nameItems := []string{}
	countItems := []int{}
	for k, v := range m {
		keys = append(keys, k)
		countItems = append(countItems, v)
	}
	sort.Strings(keys)
	for _, k := range keys {
		nameItems = append(nameItems, k)
	}
	bar := charts.NewBar()
	bar.AddXAxis(nameItems).AddYAxis("Pull Requests Opened", countItems)
	return bar
}
