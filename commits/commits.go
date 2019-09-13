package commits

import (
	"context"
	"github.com/google/go-github/github"
	"sort"
	"time"
	"log"
	"fmt"
)

func GetUserCommits(ctx context.Context, orgName string, client *github.Client, username string,
					m map[string]int, yearAgo time.Time, repoName string, repoOwner string, wC []int) {
	var list []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA: "master", 
		Author: username,
		Since: yearAgo, 
		ListOptions: github.ListOptions{PerPage: 30},
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
	getLastWeekCommits(ctx, orgName, client, username, yearAgo, repoName, repoOwner, wC)
}

func getLastWeekCommits(ctx context.Context, orgName string, client *github.Client, username string,
						yearAgo time.Time, repoName string, repoOwner string, wC []int) {
	var list1 []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA: "master", 
		ListOptions: github.ListOptions{PerPage: 30},
		Since: time.Now().AddDate(0, 0, -7),
	}
	for {
		l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
		if err != nil {
			log.Println(err)
			break
		}
		list1 = append(list1, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	var list2 []*github.RepositoryCommit
	opt = &github.CommitsListOptions{
		SHA: "master", 
		Author: username,
		ListOptions: github.ListOptions{PerPage: 30},
		Since: time.Now().AddDate(0, 0, -7),
	}
	for {
		l, resp, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, opt)
		if err != nil {
			log.Println(err)
			break
		}
		list2 = append(list2, l...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	fmt.Println("In repo", repoName, len(list1), "commits were made")
	fmt.Println("In repo", repoName, len(list2), "commits were made by the user")
	wC[0] = wC[0] + (len(list1) - len(list2))     // Difference
	wC[1] = wC[1] + len(list2)				     // Commits Made by User
	fmt.Println("Commits made by user:", wC[1], "Difference in Commits", wC[0])
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

// given map of date:conts & map of desired dates, want to return all dates in sorted order
func CommitsBase(m map[string]int, x map[string]struct{}) (map[string]struct{}) {
	for k := range m { 
		if _, ok := x[k]; !ok {
			x[k] = struct{}{} 
		}
	}
	return x
}

// want to return list of conts at those dates
func GetContsList(m map[string]int, x map[string]struct{}) ([]int, []string) {
	// for each date, if date not contained in desired dates, add value (cont) to list of all conts at those dates
	var dates []string
	var conts []int
	for k := range x {
		dates = append(dates, k)
	}
	sort.Strings(dates)
	for _, k := range dates {
		if _, ok := m[k]; !ok {
			conts = append(conts, 0) 
		} else {
			conts = append(conts, m[k])
		}
	}
	return conts, dates
}
