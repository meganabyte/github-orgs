package main

import (
	"html/template"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"github.com/meganabyte/github-orgs/repos"
	"encoding/json"
	"time"
)

type Data struct {
	User			string
	Org				string
	Token			string
	Commits         map[string]int
	Issues          map[string]int
	IssuesCommented map[string]int
	Pulls           map[string]int
	PullsMerged     map[string]int
	PullsReviewed   map[string]int
	WeekCommits		[]int
}

type Chart struct {
	X1               map[string]struct{}
	CommitsY         []int
	IssuesY          []int
	X2 				 map[string]struct{}
	PullsY           []int
	PullsMergedY     []int
	IssuesCommentedY []int
	PullsReviewedY   []int
}

func main() {
	var yearAgo = time.Now().AddDate(-1, 0, 0)
	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("assets/index.html")
		throwError(err)
		err = t.Execute(w, nil)
		throwError(err)
	}).Methods("GET")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		throwError(err)
		var d = Data{
			Commits:         make(map[string]int),
			Issues:          make(map[string]int),
			IssuesCommented: make(map[string]int),
			Pulls:           make(map[string]int),
			PullsMerged:     make(map[string]int),
			PullsReviewed:   make(map[string]int),
			WeekCommits:     make([]int, 2),
		}
		var c = Chart{
			X1: make(map[string]struct{}),
			X2: make(map[string]struct{}),
		}

		d.Org = r.FormValue("org")
		d.Token = r.FormValue("token")
		d.User = r.FormValue("user")

		ctx, client := members.Authentication(d.Token)
		list, _ := repos.GetRepos(ctx, d.Org, client)
		repos.FetchContributions(list, ctx, d.Org, client, d.User, d.Issues, d.Commits,
		d.Pulls, d.PullsMerged, d.PullsReviewed, d.IssuesCommented, d.WeekCommits, yearAgo)

		var dates1 []string
		var dates2 []string
		var dates3 []string
		c.X1 = commits.CommitsBase(d.Commits, c.X1)
		c.X1 = commits.CommitsBase(d.Issues, c.X1)
		c.X1 = commits.CommitsBase(d.Pulls, c.X1)
		c.CommitsY, _ = commits.GetContsList(d.Commits, c.X1)
		c.IssuesY, _ = commits.GetContsList(d.Issues, c.X1)
		c.PullsY, dates1 = commits.GetContsList(d.Pulls, c.X1)

		c.X2 = commits.CommitsBase(d.PullsReviewed, c.X2)
		c.X2 = commits.CommitsBase(d.PullsMerged, c.X2)
		c.PullsReviewedY, _ = commits.GetContsList(d.PullsReviewed, c.X2)
		c.PullsMergedY, dates2 = commits.GetContsList(d.PullsMerged, c.X2)
		c.IssuesCommentedY, dates3 = issues.IssuesBase(d.IssuesCommented)

		// Chart 1
		timeChart1, _ := json.Marshal(dates1)
		commitsY, _ := json.Marshal(c.CommitsY)
		issuesY, _ := json.Marshal(c.IssuesY)
		pullsY, _ := json.Marshal(c.PullsY)
	
		// Chart 2
		timeChart2, _ := json.Marshal(dates2)
		pullsReviewedY, _ := json.Marshal(c.PullsReviewedY)
		pullsMergedY, _ := json.Marshal(c.PullsMergedY)

		// Chart 3
		weeklyCommits, _ := json.Marshal(d.WeekCommits)

		// Chart 4
		timeChart4, _ := json.Marshal(dates3)
		issuesCommentedY, _ := json.Marshal(c.IssuesCommentedY)

		err = members.RenderChart(w, d.User, d.Org, timeChart1, timeChart2, timeChart4, commitsY, issuesY, pullsY,
								  pullsMergedY, pullsReviewedY, issuesCommentedY, weeklyCommits)
		throwError(err)
		
	}).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func throwError(err error) {
	if err != nil {
		log.Println(err)
		return
	}
}
