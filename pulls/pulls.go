package pulls

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
	"time"
	"fmt"
	"sync"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  m map[string]int, yearAgo time.Time, repoName string, repoOwner string) (error) {
	var wg sync.WaitGroup
	var list []*github.Issue
	opt := &github.IssueListByRepoOptions{
		Creator: username,
		State: "all",
		Since: time.Now().AddDate(0, -1, 0),
		ListOptions: github.ListOptions{PerPage: 30},
	}
	for {
		l, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repoName, opt)
		if err != nil {
			return err
		}
		list = append(list, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	wg.Add(2)
	wg.Wait()
	go func() {
		GetReviewTimes(list, m, username, client, ctx, repoOwner, repoName)
		wg.Done()
	}()
	go func() {
		GetMergedTimes(list, m, username, client, ctx, repoOwner, repoName)
		wg.Done()
	}()
	return nil
}

func GetReviewTimes(list []*github.Issue, m map[string]int, username string, client *github.Client, ctx context.Context,
				   repoOwner string, repoName string) {
	for _, issue := range list {
		num := issue.GetNumber()
		reviews, _, err := client.PullRequests.ListReviews(ctx, repoOwner, repoName, num, nil)
		if err != nil {
			return 
		}
		for _, review := range reviews {
			if review.GetUser().GetLogin() == username {
				time := review.GetSubmittedAt().Format("2006-01-02")
				fmt.Println(num, repoName, time)
			}
		}
	}
}

func GetMergedTimes(list []*github.Issue, m map[string]int, username string, client *github.Client, ctx context.Context,
					repoOwner string, repoName string) {
	for _, issue := range list {
		num := issue.GetNumber()
		pull, _, err := client.PullRequests.Get(ctx, repoOwner, repoName, num)
		if err != nil {
			return
		}
		if pull.GetMerged() && pull.GetMergedBy().GetLogin() == username {
			time := pull.GetMergedAt().Format("2006-01-02")
			fmt.Println("PR Merged at:", time)
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
