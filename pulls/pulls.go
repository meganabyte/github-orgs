package pulls

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
	"time"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  m map[string]int, yearAgo time.Time, repoName string, repoOwner string) (error) {
	/*
	var list []*github.PullRequest
	opt := &github.PullRequestListOptions{
		State: "all", 
		Base: "master",
		ListOptions: github.ListOptions{PerPage: 30},
	}
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
	GetPullsTimes(list, m, username, yearAgo)
	*/
	return nil
}

func GetPullsTimes(list []*github.PullRequest, m map[string]int, username string, yearAgo time.Time) {
	/*
	for _, pull := range list {
		time := pull.GetCreatedAt()
		if pull.GetUser().GetLogin() == username && !time.Before(yearAgo) {
			mTime := time.Format("2006-01-02")
			if val, ok := m[mTime]; !ok {
				m[mTime] = 1
			} else {
				m[mTime] = val + 1
			}
		}
	}
	*/
}

func PullsBase(m map[string]int) *charts.Bar {
	var keys []string
	nameItems := []string{}
	countItems := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		countItems = append(countItems, m[k])
		nameItems = append(nameItems, k)
	}
	bar := charts.NewBar()
	bar.AddXAxis(nameItems).AddYAxis("Pull Requests Opened", countItems)
	return bar
}
