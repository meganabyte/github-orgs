package issues

import (
	"context"
	"github.com/google/go-github/github"
)

func GetRepoIssues(ctx context.Context, client *github.Client, orgName string, repos []*github.Repository) ([]*github.Issue,
	[]*github.IssueComment, []*github.IssueEvent, error) {
	var list []*github.Issue
	var comments []*github.IssueComment
	var events []*github.IssueEvent
	for _, repo := range repos {
		repoName := repo.GetName()
		repoOwner := repo.GetOwner().GetLogin()
		opt := &github.IssueListByRepoOptions{
			State:       "all",
			ListOptions: github.ListOptions{PerPage: 30}}
		for {
			l, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repoName, opt)
			if err != nil {
				return nil, nil, nil, err
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

func GetIssuesCreated(ctx context.Context, orgName string, client *github.Client, username string, 
					  repos []*github.Repository) (map[string]int, error) {
	list, _, _, err := GetRepoIssues(ctx, client, orgName, repos)
	if err != nil {
		return nil, err
	}
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
	return m, nil
}
