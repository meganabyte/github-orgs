package issues

import (
	"context"
	"github.com/google/go-github/github"
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

// get map of date:cont, want to return list of dates and list of conts at those dates
func IssuesBase(m map[string]int) ([]int, []string) {
	// get dates in map 
	nameItems := []string{} 
	countItems := []int{} 
	for i := 0; i < 7; i++ {
		time := time.Now().AddDate(0, 0, (-7 + i)).Format("2006-01-02")
		if val, ok := m[time]; !ok {
			countItems = append(countItems, 0)
		} else {
			countItems = append(countItems, val)
		}
		nameItems = append(nameItems, time)

	}
	return countItems, nameItems
}
