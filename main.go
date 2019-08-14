package main

import (
	"fmt"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"github.com/meganabyte/github-orgs/pulls"
	"github.com/meganabyte/github-orgs/repos"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"log"
)

type User struct {
	Org string
	Token string
	Login string
}

type Data struct {
	Issues map[string]int 
	Pulls map[string]int 
	Commits map[string]int	
}

var u User

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("assets/index.html")
		err = t.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		u.Org = r.FormValue("org")
		u.Token = r.FormValue("token")
		u.Login = r.FormValue("user")

		t, err := template.ParseFiles("assets/report.html")
		if err != nil {
			fmt.Println(err)
			return
		}

		ctx, client := members.Authentication(u.Token)
		users, _ := members.GetMembers(ctx, u.Org, client)
		if members.ContainsUser(users, u.Login) {
			fmt.Println("Loading report for", u.Login, "...")
		}

		repos, _ := repos.GetRepos(ctx, u.Org, client)
		i, _ := issues.GetIssuesCreated(ctx, u.Org, client, u.Login, repos)
		c, _ := commits.GetUserCommits(ctx, u.Org, client, u.Login, repos)
		p, _ := pulls.GetUserPulls(ctx, u.Org, client, u.Login, repos)
		d := Data {
			Issues: i,
			Pulls: p,
			Commits: c,
		}

		err = t.Execute(w, d)

	}).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
