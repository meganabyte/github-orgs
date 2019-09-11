package commits

import (
	"context"
	"github.com/google/go-github/github"
	"sort"
	"time"
	"log"
)

func GetUserCommits(ctx context.Context, orgName string, client *github.Client, username string,
					m map[string]int, yearAgo time.Time, repoName string, repoOwner string) {
	var list []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA: "master", 
		Author: username, 
		ListOptions: github.ListOptions{PerPage: 30},
		Since: yearAgo,
		//Until: time.Now().AddDate(0, 0, -7),
	}
	for {
		l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
		if err != nil {
			log.Println(err)
			return
		}
		list = append(list, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	getCommitTimes(list, m)
}

/*
func getLastWeekCommits(ctx context.Context, orgName string, client *github.Client, username string,
						m map[string]int, yearAgo time.Time, repoName string, repoOwner string) {
	var list []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA: "master", 
		ListOptions: github.ListOptions{PerPage: 30},
		Since: time.Now().AddDate(0, 0, -7),
	}
	for {
		l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
		if err != nil {
			log.Println(err)
			return
		}
		list = append(list, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	getCommitTimes(list, m)
}
*/

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

func CommitsBase(m map[string]int, x map[string]struct{}) (map[string]struct{}, []int) {
	var keys []string
	countItems := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		countItems = append(countItems, m[k])
		if _, ok := x[k]; !ok {
			x[k] = struct{}{}
		}
	}
	return x, countItems
}
