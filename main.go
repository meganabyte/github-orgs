package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"github.com/meganabyte/github-orgs/pulls"
	"github.com/meganabyte/github-orgs/repos"
	"fmt"
	"time"
)

type User struct {
	Org   string
	Token string
	Login string
}

type Data struct {
	User string
	Org string
	Issues  map[string]int
	Pulls   map[string]int
	Commits map[string]int
}

var (
	u User
	d = Data{
		User: "",
		Org: "",
		Issues:  make(map[string]int),
		Pulls:   make(map[string]int),
		Commits: make(map[string]int),
	}
)

// assumes there is a DynamoDB table named UserData

func main() {
	var yearAgo = time.Now().AddDate(-1, 0, 0)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
	}))
	svc := dynamodb.New(sess)
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
		d.User = u.Login
		d.Org = u.Org

		params := &dynamodb.GetItemInput{
			TableName: aws.String("UserData"),
			Key: map[string]*dynamodb.AttributeValue {
				"User": {
					S: aws.String(d.User),
				},
				"Org": {
					S: aws.String(d.Org),
				},
			},
		}
		result, err := svc.GetItem(params)
		if err != nil {
			log.Println(err)
			return
		}

		if len(result.Item) == 0 {
			// compute user data
			ctx, client := members.Authentication(u.Token)
			repos, _ := repos.GetRepos(ctx, u.Org, client)
			err = issues.GetIssuesCreated(ctx, u.Org, client, u.Login, repos, d.Issues, yearAgo)
			if err != nil {
				log.Println(err)
				return
			}
			err = commits.GetUserCommits(ctx, u.Org, client, u.Login, repos, d.Commits, yearAgo)
			if err != nil {
				log.Println(err)
				return
			}
			err = pulls.GetUserPulls(ctx, u.Org, client, u.Login, repos, d.Pulls, yearAgo)
			if err != nil {
				log.Println(err)
				return
			}

			// creates item for entered user
			av, err := dynamodbattribute.MarshalMap(d)
			if err != nil {
				log.Println(err)
				return
			}
			input := &dynamodb.PutItemInput{
				Item:      av,
				TableName: aws.String("UserData"),
			}
			_, err = svc.PutItem(input)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println(d.Commits, d.Pulls, d.Issues)

		} else {
			item := Data{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &item)
			if err != nil {
				log.Println(err)
			}
			d.Commits = item.Commits
			d.Issues = item.Issues
			d.Pulls = item.Pulls
			fmt.Println(d.Commits, d.Pulls, d.Issues)
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
