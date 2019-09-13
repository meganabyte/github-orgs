package members

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"context"
	"io"
	"net/http"
)

func Authentication(token string) (context.Context, *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return ctx, client
}

func RenderChart(w http.ResponseWriter, login string, org string, timeChart1 string, timeChart2 string, timeChart4 string, 
				 commitsY []byte, issuesY []byte, pullsY []byte, pullsMergedY []byte, pullsReviewedY []byte, issuesCommentedY []byte, 
				 weeklyCommits[]byte) (error) {
	_, err := io.WriteString(w, `<!DOCTYPE html>
		<html>
		  <head>
			<title>Contributions Report</title>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.5.0/Chart.min.js"></script>
			<link rel="stylesheet" type="text/css" href="../static/style.css">
		  </head>
		  <body>
			  <h1>`+login+`'s contributions to `+org+`</h1>
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
				var time3 = `+string(timeChart4)+`
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
							'`+login+`'
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
			return err
		}
		return nil
}
