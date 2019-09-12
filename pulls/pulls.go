package pulls

import (
	"context"
	"github.com/google/go-github/github"
	"time"
	"fmt"
	"log"
	//"sync"
)

func GetUserPulls(ctx context.Context, orgName string, client *github.Client, username string,
				  pM map[string]int, pR map[string]int, iC map[string]int, repoName string, repoOwner string) {
	//var wg sync.WaitGroup
	var list []*github.Issue
	opt := &github.IssueListByRepoOptions{
		State: "all",
		Since: time.Now().AddDate(0, 0, -7),
		ListOptions: github.ListOptions{PerPage: 30},
	}
	for {
		l, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repoName, opt)
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
	for _, issue := range list {
		num := issue.GetNumber()
		if issue.IsPullRequest() {
			//wg.Add(2)
			//wg.Wait()
			//go func() {
				getReviewTimes(num, username, pR, client, ctx, repoOwner, repoName)
				//wg.Done()
			//}()
			//go func() {
				getMergedTimes(num, username, pM, client, ctx, repoOwner, repoName)
				//wg.Done()
			//}()
		}
		getIssueCommentTimes(num, username, iC, client, ctx, repoOwner, repoName)
	}
}

func getReviewTimes(num int, username string, m map[string]int, client *github.Client, ctx context.Context,
				   repoOwner string, repoName string) {
	reviews, _, err := client.PullRequests.ListReviews(ctx, repoOwner, repoName, num, nil)
	if err != nil {
		log.Println(err)
		return 
	}
	for _, review := range reviews {
		if review.GetUser().GetLogin() == username {
			time := review.GetSubmittedAt().Format("2006-01-02")
			fmt.Println("PR #", num, "reviewed at", time, "in repo", repoName)
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
}

func getMergedTimes(num int, username string, m map[string]int, client *github.Client, ctx context.Context,
					repoOwner string, repoName string) {
	pull, _, err := client.PullRequests.Get(ctx, repoOwner, repoName, num)
	if err != nil {
		log.Println(err)
		return
	}
	if pull.GetMerged() && pull.GetMergedBy().GetLogin() == username {
		time := pull.GetMergedAt().Format("2006-01-02")
		fmt.Println("PR #", num, "merged at", time, "in repo", repoName)
		if val, ok := m[time]; !ok {
			m[time] = 1
		} else {
			m[time] = val + 1
		}
	}
}
func getIssueCommentTimes(num int, username string, m map[string]int, client *github.Client, ctx context.Context,
						  repoOwner string, repoName string) {
	comments, _, err := client.Issues.ListComments(ctx, repoOwner, repoName, num, nil)
	if err != nil {
		log.Println(err)
		return 
	}
	for _, comment := range comments {
		if comment.GetUser().GetLogin() == username {
			time := comment.GetCreatedAt().Format("2006-01-02")
			fmt.Println("Issue #", num, "comment made at", time, "in repo", repoName)
			if val, ok := m[time]; !ok {
				m[time] = 1
			} else {
				m[time] = val + 1
			}
		}
	}
}
