package members

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"context"
)
func GetMembers(ctx context.Context, orgName string, client *github.Client) ([]*github.User, error) {
	var list []*github.User
	opt := &github.ListMembersOptions{ListOptions: github.ListOptions{PerPage: 30}}
	for {
		members, resp, err := client.Organizations.ListMembers(ctx, orgName, opt)
		if err != nil {
			return nil, err
		}
		list = append(list, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return list, nil
}

func Authentication(token string) (context.Context, *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return ctx, client
}

func ContainsUser(users []*github.User, login string) bool {
	for _, user := range users {
		username := user.GetLogin()
		if username == login {
			return true
		}
	}
	return false
}
