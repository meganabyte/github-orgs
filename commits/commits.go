package commits

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/chenjiandongx/go-echarts/charts"
	"sort"
	"time"
	"fmt"
)

func GetUserCommits(ctx context.Context, orgName string, client *github.Client, username string,
					m map[string]int, yearAgo time.Time, repoName string, repoOwner string) (error) {
	start := time.Now()
	var list []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA: "master", 
		Author: username, 
		ListOptions: github.ListOptions{PerPage: 30},
		Since: yearAgo,
	}
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
	getCommitTimes(list, m)
	fmt.Println("Finished fetching commits after ", time.Since(start))
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
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		countItems = append(countItems, m[k])
		nameItems = append(nameItems, k)
	}
	bar := charts.NewBar()
	bar.AddXAxis(nameItems).AddYAxis("Commits Created", countItems)
	return bar
}
