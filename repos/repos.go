package repos

import (
	"github.com/google/go-github/github"
	"context"
)

func GetRepos(ctx context.Context, orgName string, client *github.Client) ([]*github.Repository, error) {
	var list []*github.Repository
	opt := &github.RepositoryListByOrgOptions{Type: "sources", ListOptions: github.ListOptions{PerPage: 30}}
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, orgName, opt)
		if err != nil {
			return nil, err
		}
		list = append(list, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return list, nil
}
