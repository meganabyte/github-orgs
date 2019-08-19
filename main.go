package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"github.com/meganabyte/github-orgs/pulls"
	"github.com/meganabyte/github-orgs/repos"
)

type User struct {
	Org   string
	Token string
	Login string
}

type Data struct {
	Issues  map[string]int
	Pulls   map[string]int
	Commits map[string]int
}

var (
	u User
	d = Data{
		Issues:  make(map[string]int),
		Pulls:   make(map[string]int),
		Commits: make(map[string]int),
	}
)

func main() {
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("assets/index.html")
		if err != nil {
			log.Println(err)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Println(err)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}
		u.Org = r.FormValue("org")
		u.Token = r.FormValue("token")
		u.Login = r.FormValue("user")

		ctx, client := members.Authentication(u.Token)
		repos, _ := repos.GetRepos(ctx, u.Org, client)

		err = issues.GetIssuesCreated(ctx, u.Org, client, u.Login, repos, d.Issues)
		if err != nil {
			log.Println(err)
			return
		}
		err = commits.GetUserCommits(ctx, u.Org, client, u.Login, repos, d.Commits)
		if err != nil {
			log.Println(err)
			return
		}
		err = pulls.GetUserPulls(ctx, u.Org, client, u.Login, repos, d.Pulls)
		if err != nil {
			log.Println(err)
			return
		}
		bar := commits.CommitsBase(d.Commits)
		bar.Overlap(issues.IssuesBase(d.Issues), pulls.PullsBase(d.Pulls))
		bar.Title = u.Login
		f, err := os.Create("bar.html")
		if err != nil {
			log.Println(err)
			return
		}
		err = bar.Render(w, f)
		if err != nil {
			log.Println(err)
			return
		}

	}).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
