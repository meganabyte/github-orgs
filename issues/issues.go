package issues

import (
	"repos"
	"github.com/google/go-github/github"
	"context"
	"fmt"
)

func GetRepoIssues(ctx context.Context, client *github.Client, orgName string) ([]*github.Issue, 
	[]*github.IssueComment, 
	[]*github.IssueEvent, error) {
	repos, _ := repos.GetRepos(ctx, orgName, client)
	var list []*github.Issue
	var comments []*github.IssueComment
	var events []*github.IssueEvent
	for _, repo := range repos {
		repoName := repo.GetName()
		repoOwner := repo.GetOwner().GetLogin()
		opt := &github.IssueListByRepoOptions{
			State: "all", 
			ListOptions: github.ListOptions{PerPage: 30}}
		for {
			l, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repoName, opt)
			if err != nil {
				return nil, nil, nil, err
			}
			for _, issue := range l {
				if issue.GetComments() != 0 {
					num := issue.GetNumber()
					opt := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 30}}
					c, _, err := client.Issues.ListComments(ctx, repoOwner, repoName, num, opt)
					e, _, err := client.Issues.ListIssueEvents(ctx, repoOwner, repoName, num, nil)
					if err != nil {
						fmt.Println(err)
						return nil, nil, nil, err
					}
					comments = append(comments, c...)
					events = append(events, e...)
				}
			}
			list = append(list, l...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}
	return list, comments, events, nil
}

func GetIssuesCreated(ctx context.Context, orgName string, client *github.Client, username string) (map[string]int) {
	list, _, _ , _ := GetRepoIssues(ctx, client, orgName)
	m := make(map[string]int)
	for _, issue := range list {
		if issue.GetUser().GetLogin() == username {
			time := issue.GetCreatedAt().Format("2006-01-02")
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
	return m
}

func GetIssueComments(ctx context.Context, orgName string, client *github.Client, username string) (map[string]int) {
	_, list, _, _ := GetRepoIssues(ctx, client, orgName)
	m := make(map[string]int)
	for _, comment := range list {
		if comment.GetUser().GetLogin() == username {
			time := comment.GetCreatedAt().Format("2006-01-02")
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
	return m
}

func GetIssueEvents(ctx context.Context, orgName string, client *github.Client, username string) (map[string]int) {
	_, _, list, _ := GetRepoIssues(ctx, client, orgName)
	m := make(map[string]int)
	for _, event := range list {
		if *event.Event == "closed" && event.Actor.GetLogin() == username {
			time := event.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00")
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
	return m
}
