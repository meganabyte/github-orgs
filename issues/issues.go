package issues

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
	"time"
	"log"
)

func GetRepoIssues(ctx context.Context, client *github.Client, orgName string, repoName string, 
					repoOwner string, username string, yearAgo time.Time) ([]*github.Issue, error) {
	var list []*github.Issue
	opt := &github.IssueListByRepoOptions{
		Creator: username,
		State: "all",
		Since: yearAgo,
		ListOptions: github.ListOptions{PerPage: 30},
	}
	for {
		l, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repoName, opt)
		if err != nil {
			return nil, err
		}
		list = append(list, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return list, nil
}

func GetIssuesCreated(ctx context.Context, orgName string, client *github.Client, username string, i map[string]int,
					   p map[string]int, yearAgo time.Time, repoName string, repoOwner string) {
	list, err := GetRepoIssues(ctx, client, orgName, repoName, repoOwner, username, yearAgo)
	if err != nil {
		log.Println(err)
		return 
	}
	for _, issue := range list {
		time := issue.GetCreatedAt().Format("2006-01-02")
		if !issue.IsPullRequest() {
			if val, ok := i[time]; !ok {
				i[time] = 1
			} else {
				i[time] = val + 1
			}
		} else {
			if issue.GetUser().GetLogin() == username {
				if val, ok := p[time]; !ok {
					p[time] = 1
				} else {
					p[time] = val + 1
				}
			}
		}
	}
}

func IssuesBase(m map[string]int) *charts.Bar {
	var keys []string
	nameItems := []string{} // x axis
	countItems := []int{} // y axis
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		countItems = append(countItems, m[k])
		nameItems = append(nameItems, k)
	}
	bar := charts.NewBar()
	bar.AddXAxis(nameItems).AddYAxis("Issues Opened", countItems)
	return bar
}
