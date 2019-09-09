package pulls

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
	"time"
	"fmt"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  m map[string]int, yearAgo time.Time, repoName string, repoOwner string) (error) {
	start := time.Now()
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
	fmt.Println("Finished fetching pulls after ", time.Since(start))
	GetPullsTimes(list, m, username, yearAgo)
	return nil
}

func GetPullsTimes(list []*github.PullRequest, m map[string]int, username string, yearAgo time.Time) {
	start := time.Now()
	for _, pull := range list {
		if pull.GetUser().GetLogin() == username {
			time := pull.GetCreatedAt()
			if !time.Before(yearAgo) {
				mTime := time.Format("2006-01-02")
				if val, ok := m[mTime]; !ok {
					m[mTime] = 1
				} else {
					m[mTime] = val + 1
				}
			}
		}
	}
	fmt.Println("Finished processing pulls after ", time.Since(start))
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
