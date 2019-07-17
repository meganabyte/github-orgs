package members

import (
	"github.com/google/go-github/github"
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
