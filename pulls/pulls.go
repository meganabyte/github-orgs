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
				  pM map[string]int, pR map[string]int, repoName string, repoOwner string) (error) {
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
	for _, issue := range list {
		if issue.IsPullRequest() {
			num := issue.GetNumber()
			getReviewTimes(num, username, pR, client, ctx, repoOwner, repoName)
			getMergedTimes(num, username, pM, client, ctx, repoOwner, repoName)
		}
	}
	return nil
}

func getReviewTimes(num int, username string, m map[string]int, client *github.Client, ctx context.Context,
				   repoOwner string, repoName string) {
	reviews, _, err := client.PullRequests.ListReviews(ctx, repoOwner, repoName, num, nil)
	if err != nil {
		return 
	}
	for _, review := range reviews {
		if review.GetUser().GetLogin() == username {
			time := review.GetSubmittedAt().Format("2006-01-02")
			fmt.Println("PR #", num, "reviewed at", time)
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
}

func getMergedTimes(num int, username string, m map[string]int, client *github.Client, ctx context.Context,
					repoOwner string, repoName string) {
	pull, _, err := client.PullRequests.Get(ctx, repoOwner, repoName, num)
	if err != nil {
		return
	}
	if pull.GetMerged() && pull.GetMergedBy().GetLogin() == username {
		time := pull.GetMergedAt().Format("2006-01-02")
		fmt.Println("PR #", num, "merged at", time)
		if val, ok := m[time]; !ok {
			m[time] = 1
		} else {
			m[time] = val + 1
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
