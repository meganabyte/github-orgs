package repos

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/pulls"
	"time"
	"fmt"
	//"sync"
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

func FetchContributions(repos []*github.Repository, ctx context.Context, orgName string, client *github.Client, username string,
				i map[string]int, c map[string]int, p map[string]int, pM map[string]int, pR map[string]int, iC map[string]int, 
				cD int, wC int, yearAgo time.Time) {
	//var wg sync.WaitGroup					
	start := time.Now()
	for _, repo := range repos {
		if repo.GetSize() != 0 {
			repoName := repo.GetName()
			repoOwner := repo.GetOwner().GetLogin()
			//wg.Add(3)
			//wg.Wait()
			//go func() {
				issues.GetIssuesCreated(ctx, orgName, client, username, i, p, yearAgo, repoName, repoOwner)
				//wg.Done()
			//}()
			//go func() {
				commits.GetUserCommits(ctx, orgName, client, username, c, yearAgo, repoName, repoOwner, cD, wC)
				//wg.Done()
			//}()
			//go func() {
				pulls.GetUserPulls(ctx, orgName, client, username, pM, pR, iC, repoName, repoOwner)
				//wg.Done()
			//}()
		}
	}
	fmt.Println("Finished fetching cont after ", time.Since(start))
}
