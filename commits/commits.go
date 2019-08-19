package commits

import (
	"context"
	"sort"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
)

func GetUserCommits(ctx context.Context, orgName string, client *github.Client, username string,
					repos []*github.Repository, m map[string]int) (error) {
	var list []*github.RepositoryCommit
	for _, repo := range repos {
		if repo.GetSize() != 0 {
			repoName := repo.GetName()
			repoOwner := repo.GetOwner().GetLogin()
			opt := &github.CommitsListOptions{SHA: "master", Author: username, ListOptions: github.ListOptions{PerPage: 30}}
			for {
				l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
				if err != nil {
					return err
				}
				list = append(list, l...)
				if resp.NextPage == 0 {
					break
				}
				opt.Page = resp.NextPage
			}
		}
	}
	getCommitTimes(list, m)
	return nil
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

func CommitsBase(m map[string]int) *charts.Bar {
	var keys []string
	nameItems := []string{}
	countItems := []int{}
	for k, v := range m {
		keys = append(keys, k)
		countItems = append(countItems, v)
	}
	sort.Strings(keys)
	for _, k := range keys {
		nameItems = append(nameItems, k)
	}
	bar := charts.NewBar()
	bar.AddXAxis(nameItems).AddYAxis("Commits Created", countItems)
	return bar
}
