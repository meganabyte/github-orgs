package pulls

import (
	"context"
	"github.com/google/go-github/github"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  repos []*github.Repository) (map[string]int, error) {
	var list []*github.PullRequest
	for _, repo := range repos {
		repoName := repo.GetName()
		repoOwner := repo.GetOwner().GetLogin()
		opt := &github.PullRequestListOptions{State: "all", Base: "master", ListOptions: github.ListOptions{PerPage: 30}}
		for {
			l, resp, err := client.PullRequests.List(ctx, repoOwner, repoName, opt)
			if err != nil {
				return nil, err
			}
			list = append(list, l...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}
	m := make(map[string]int)
	GetPullsTimes(list, m, username)
	return m, nil
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
