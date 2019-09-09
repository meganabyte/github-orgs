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

func GetPullsTimes(pull *github.Issue, m map[string]int, username string, client *github.Client, ctx context.Context,
				   repoOwner string, repoName string) {
	num := pull.GetNumber()
	reviews, _, err := client.PullRequests.ListReviews(ctx, repoOwner, repoName, num, nil)
	if err != nil {
		return 
	}
	for _, review := range reviews {
		if review.GetUser().GetLogin() == username {
			time := pull.GetCreatedAt().Format("2006-01-02")
			fmt.Println(num, repoName, time)
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
