package main

import (
	"context"
	//"flag"
	"fmt"
	"os"
	"github.com/google/go-github/github"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"github.com/meganabyte/github-orgs/pulls"
	"golang.org/x/oauth2"
)

func main() {
	/*flag.Parse()
	args := flag.Args()
	org := args[0]
	if len(args) < 2 {
		fmt.Println("go run main <Organization Name> <OAUTH token>")
		os.Exit(1)
	}
	token := args[1]
	*/
	org := os.Getenv("ORG_NAME")
	token := os.Getenv("OAUTH_TOKEN")
	ctx, client := authentication(token)
	users, _ := members.GetMembers(ctx, org, client)
	for _, user := range users {
		username := user.GetLogin()
		c, _ := commits.GetUserCommits(ctx, org, client, username)
		p, _ := pulls.GetUserPulls(ctx, org, client, username)
		i := issues.GetIssueTimes(ctx, org, client, username)
		fmt.Println("Issues created by", username, ":", i)
		fmt.Println("Commits created by", username, ": ", c)
		fmt.Println("Pulls created by", username, ": ", p)
	}
}

func authentication(token string) (context.Context, *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return ctx, client
}
