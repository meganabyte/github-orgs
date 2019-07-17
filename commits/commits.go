package commits

import (
	"repos"
	"github.com/google/go-github/github"
	"context"
)

func GetUserCommits(ctx context.Context, orgName string, client *github.Client, username string) (map[string]int, error) {
	repos, _ := repos.GetRepos(ctx, orgName, client)
	var list []*github.RepositoryCommit
	for _, repo := range repos {
		if repo.GetSize() != 0 {
			repoName := repo.GetName()
			repoOwner := repo.GetOwner().GetLogin()
			opt := &github.CommitsListOptions{SHA: "master", Author: username, ListOptions: github.ListOptions{PerPage: 30}}
			for {
				l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
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
	}
	m := make(map[string]int)
	getCommitTimes(list, m)
	return m, nil
}

func getCommitTimes(list []*github.RepositoryCommit, m map[string]int) {
	for _, commit := range list {
		author := commit.Commit.GetAuthor()
		time := author.GetDate().Format("2006-01-02")
		if val, ok := m[time]; !ok {
			m[time] = 1
		} else {
			m[time] = val + 1
		}
	}
}
