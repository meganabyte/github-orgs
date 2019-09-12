package main

import (
	"html/template"
	"log"
	"net/http"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/meganabyte/github-orgs/commits"
	"github.com/meganabyte/github-orgs/issues"
	"github.com/meganabyte/github-orgs/members"
	"io"
	"time"
	"github.com/meganabyte/github-orgs/repos"
	"encoding/json"
	"fmt"
)

type User struct {
	Org   string
	Token string
	Login string
}

type Data struct {
	User			string
	Org				string
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
	IssuesCommentedY []int
	PullsY           []int
	PullsMergedY     []int
	PullsReviewedY   []int
}

var (
	u User
	d = Data{
		Commits:         make(map[string]int),
		Issues:          make(map[string]int),
		IssuesCommented: make(map[string]int),
		Pulls:           make(map[string]int),
		PullsMerged:     make(map[string]int),
		PullsReviewed:   make(map[string]int),
		WeekCommits:     make([]int, 2),
	}
	c = Chart{
		X1: make(map[string]struct{}),
		X2: make(map[string]struct{}),
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
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

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
			Key: map[string]*dynamodb.AttributeValue{
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
		// if this item doesn't exist
		if len(result.Item) == 0 {
			// compute user data
			ctx, client := members.Authentication(u.Token)
			list, _ := repos.GetRepos(ctx, u.Org, client)
			repos.FetchContributions(list, ctx, u.Org, client, u.Login, d.Issues, d.Commits,
			d.Pulls, d.PullsMerged, d.PullsReviewed, d.IssuesCommented, d.WeekCommits, yearAgo)

			fmt.Println(d.WeekCommits)

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

		} else {
			item := Data{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &item)
			if err != nil {
				log.Println(err)
			}
			d.Commits = item.Commits
			d.Issues = item.Issues
			d.Pulls = item.Pulls
			d.PullsMerged = item.PullsMerged
			d.PullsReviewed = item.PullsReviewed
			d.IssuesCommented = item.IssuesCommented
			d.WeekCommits = item.WeekCommits
		}

		var dates1 []string
		var dates2 []string
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
		issuesCommented, dates3 := issues.IssuesBase(d.IssuesCommented)

		timeChart1, _ := json.Marshal(dates1)
		commitsY, _ := json.Marshal(c.CommitsY)
		issuesY, _ := json.Marshal(c.IssuesY)
		pullsY, _ := json.Marshal(c.PullsY)
	
		timeChart2, _ := json.Marshal(dates2)
		pullsReviewedY, _ := json.Marshal(c.PullsReviewedY)
		pullsMergedY, _ := json.Marshal(c.PullsMergedY)
		issuesCommentedY, _ := json.Marshal(issuesCommented)

		timeChart3, _ := json.Marshal(dates3)

		weeklyCommits, _ := json.Marshal(d.WeekCommits)

		_, err = io.WriteString(w, `<!DOCTYPE html>
		<html>
		  <head>
			<title>Contributions Report</title>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.5.0/Chart.min.js"></script>
			<link rel="stylesheet" type="text/css" href="../static/style.css">
		  </head>
		  <body>
			  <h1>`+u.Login+`'s contributions to `+u.Org+`</h1>
			<div class="wrapper"style="width:65%" >
			  <canvas id="chart1" width="1300" height="800" style="margin-bottom:3em"></canvas>
			</div>
			<div style="width:40%;float: right;" class="col-xs-6">
			  <canvas id="chart3" width="1200" height="800" style="float:right;margin-right:5em;margin-top:5em"></canvas>
			</div>
			<div style="width:50%;float:left;" class="col-xs-6"">
			  <canvas id="chart4" width="1200" height="800" style="float:left;margin-left:7em;margin-bottom:5em;margin-top:3em">
			  </canvas>
			</div> 
			<div style="width:50%" class="wrapper2">
			  <canvas id="chart2" width="1200" height="800"></canvas>
			</div>
		
			<script type="text/javascript">
				var time1 = `+string(timeChart1)+`
				var time2 = `+string(timeChart2)+`
				var time3 = `+string(timeChart3)+`
				var commits = `+string(commitsY)+`
				var issues = `+string(issuesY)+`
				var pulls = `+string(pullsY)+`
				var pullsMerged = `+string(pullsMergedY)+`
				var pullsReviewed = `+string(pullsReviewedY)+`
				var issuesCommented = `+string(issuesCommentedY)+`
				var weeklyCommits = `+string(weeklyCommits)+`
				var ctx1 = document.getElementById("chart1");
				var chart1 = new Chart(ctx1, {
					type: 'bar',
					data: {
						labels: time1,
						datasets: [
						{ 
							data: commits,
							label: "Commits Made",
							backgroundColor: "#f6b93b",
							borderColor: "#f6b93b",
							fill: true
						},
						{ 
							data: issues,
							label: "Issues Opened",
							borderColor: "#e55039",
							backgroundColor: "#e55039",
							fill: true
						},
						{ 
							data: pulls,
							label: "PRs Opened",
							borderColor: "#4a69bd",
							backgroundColor: "#4a69bd",
							fill: true
						}
						]
					},
					options: {
						tooltips: {
							mode: 'index',
							intersect: false
						},
						responsive: true, 
						scales: {
							yAxes: [{
								stacked:true,
								ticks: {
									beginAtZero: true
								}
							}]
						},
						title: {
							display: true,
							text: 'Contributions over the last year',
							fontFamily: "'Roboto', sans-serif",
							fontSize: 17
						}
					}
				});
				var ctx2 = document.getElementById("chart2");
				var chart2 = new Chart(ctx2, {
					type: 'bar',
					data: {
						labels: time2,
						datasets: [
							{ 
								data: pullsMerged,
								label: "PRs Merged",
								backgroundColor: "#38ada9",
								borderColor: "#38ada9",
								fill: true
							},
							{ 
								data: pullsReviewed,
								label: "PRs Reviewed",
								borderColor: "#82ccdd",
								backgroundColor: "#82ccdd",
								fill: true
							}
						]
					},
					options: {
						tooltips: {
							mode: 'index',
							intersect: false
						},
						responsive: true, 
						scales: {
							yAxes: [{
								stacked: true,
								ticks: {
									beginAtZero: true
								}
							}]
						},
						title: {
							display: true,
							text: 'PRs merged & reviewed in the last week',
							fontFamily: "'Roboto', sans-serif",
							fontSize: 15
						}
					}
				});
				var ctx3 = document.getElementById("chart3");
				var chart3 = new Chart(ctx3, {
					type: 'doughnut',
					data : {
						datasets: [{
						data: weeklyCommits, 
						fill: true,
						backgroundColor: ["#78e08f", "#b8e994"],
						}],
		
						labels: [
							'others',
							'`+u.Login+`'
						]
					},
					options: {
						title: {
							display: true,
							text: 'Commit contribution in the last week',
							fontFamily: "'Roboto', sans-serif",
							fontSize: 15
						}
					}
				})
				
				var ctx4 = document.getElementById("chart4");
				var chart4 = new Chart(ctx4, {
					type: 'bar',
					data: {
						labels: time3,
						datasets: [
						{ 
							data: issuesCommented,
							label: "comments",
							borderColor: "#ff5252",
							backgroundColor: "#ff5252",
							fill: true
						}
						]
					},
					options: {
						tooltips: {
							mode: 'index',
							intersect: false
						},
						responsive: true, 
						scales: {
							yAxes: [{
								ticks: {
									beginAtZero: true
								}
							}]
						},
						title: {
							display: true,
							text: 'Issue Comments in the last week',
							fontFamily: "'Roboto', sans-serif",
							fontSize: 15
						}
					}
				});
			</script>
			
		  </body>
		</html>
		`)

		if err != nil {
			log.Println(err)
			return
		}

	}).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
