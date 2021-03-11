package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	chart "github.com/wcharczuk/go-chart/v2"
)

const defaultSleepInterval = "300"
const numItems = "6000"

type Login struct {
	Login string `json:"login"`
}

type Answer struct {
	Author    Login  `json:"author`
	CreatedAt string `json:"createdAt"`
	// Body      string `json:"body"`
}

type CommentNode struct {
	CreatedAt string `json:"createdAt`
	Author    Login  `json:"author"`
}

type CommentEdge struct {
	Node CommentNode `json:"node"`
}

type Comments struct {
	Edges []CommentEdge `json:"edges"`
}

type DiscussionNode struct {
	Answer         Answer   `json:"answer"`
	AnswerChosenAt string   `json:"answerChosenAt"`
	AnswerChosenBy Login    `json:"answerChosenBy"`
	Comments       Comments `json:"comments"`
	CreatedAt      string   `json:"createdAt"`
	Author         Login    `json:"author"`
	Title          string   `json:"title"`
	URL            string   `json:"url"`
}

type DiscussionEdge struct {
	Node DiscussionNode `json:"node"`
}

type PageInfo struct {
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage`
}

type Discussions struct {
	TotalCount int              `json:"totalCount"`
	PageInfo   PageInfo         `json:"pageInfo"`
	Edges      []DiscussionEdge `json:"edges"`
}

type Repository struct {
	Discussions Discussions `json:"discussions"`
}

type Data struct {
	Repository Repository `json:"repository"`
}

type GraphQL struct {
	Data Data `json:"data"`
}

type Analysis struct {
	URL                string
	Title              string
	Author             string
	CreatedAt          string
	AnsweredAt         string
	AnsweredBy         string
	ChosenAt           string
	ChosenBy           string
	NumComments        int
	ProlificCommentors string
	ProlificCount      int
}

func logJSON(data []byte) error {
	var i interface{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return err
	}
	t, err := json.MarshalIndent(&i, "", "  ")
	if err != nil {
		return err
	}
	log.Infof("JSON: ", string(t))
	return nil
}

func getRepositoryData(URL string, token string, query string) (GraphQL, error) {
	var jsonData = []byte(query)
	var graphQLObject GraphQL
	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return graphQLObject, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))
	request.Header.Set("GraphQL-Features", "discussions_api")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return graphQLObject, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return graphQLObject, err
	}
	err = json.Unmarshal(body, &graphQLObject)
	if err != nil {
		return graphQLObject, err
	}

	// The raw JSON, formatted nicely
	logJSON(body)

	// fmt.Println("############################")

	return graphQLObject, nil
}

func formatGraphQLStatement(input string) string {
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\t", " ")
	// input = strings.Replace(input, "    ", " ", -1)
	return input
}

func epoch(s string) int64 {
	layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Fatal(err)
	}
	return t.Unix()
}

func human(d int64) string {
	h := time.Unix(d, 0)
	layout := "2006-01-02T15:04:05Z"
	return h.Format(layout)
}

func amountOfTime(t int64) string {
	s := []string{}
	if t >= 60*60*24*365 {
		years := t / (60 * 60 * 24 * 365)
		t = t % (60 * 60 * 24 * 365)
		s = []string{fmt.Sprintf("%d year", years)}
	}
	if t >= 60*60*24*30 {
		months := t / (60 * 60 * 24 * 30)
		t = t % (60 * 60 * 24 * 30)
		s = append(s, fmt.Sprintf("%d month", months))
	}
	if t >= 60*60*24 {
		days := t / (60 * 60 * 24)
		t = t % (60 * 60 * 24)
		s = append(s, fmt.Sprintf("%d day", days))
	}
	if t >= 60*60 {
		hours := t / (60 * 60)
		t = t % (60 * 60)
		s = append(s, fmt.Sprintf("%d hour", hours))
	}
	if t >= 60 {
		minutes := t / (60)
		t = t % (60)
		s = append(s, fmt.Sprintf("%d minute", minutes))
	}
	if t > 0 {
		s = append(s, fmt.Sprintf("%d second", t))
	}

	return strings.Join(s, " ")
}

func main() {
	numItemsInt, err := strconv.Atoi(numItems)
	if err != nil {
		log.Fatalf("Unable to convert %s to integer", numItems)
	}

	fileName := "./results.json"
	var analyses []Analysis

	if _, err := os.Stat(fileName); err != nil {
		log.Infof("File %s does not exist. Creating it ...", fileName)
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			log.Fatal("No GITHUB_TOKEN set.")
		}

		githubAccount := os.Getenv("GITHUB_ACCOUNT")
		if githubAccount == "" {
			log.Fatal("No GITHUB_ACCOUNT set.")
		}

		githubRepo := os.Getenv("GITHUB_REPO")
		if githubRepo == "" {
			log.Fatal("No GITHUB_REPO set.")
		}

		sleepInterval := os.Getenv("GITHUB_QUERY_SLEEP_INTERVAL")
		if sleepInterval == "" {
			sleepInterval = defaultSleepInterval
		}

		var graphQLObjects []GraphQL

		totalCount := 0
		maxCount := numItemsInt
		initialCursor := "first: 100"
		for cursor, currCount, nextPage := initialCursor, 0, false; currCount < maxCount && (cursor == initialCursor || nextPage); {
			if cursor != initialCursor {
				log.Infof("Getting %s/%s %d/%d discussions at cursor %s", githubAccount, githubRepo, currCount, totalCount, cursor)
			} else {
				log.Infof("Getting %s/%s %d discussions", githubAccount, githubRepo, maxCount)
			}
			// Use this for vercel/next.js:
			//			discussions(%s categoryId: \"MDE4OkRpc2N1c3Npb25DYXRlZ29yeTY1Mjg1\" orderBy: {field: CREATED_AT, direction: DESC}) {

			query := fmt.Sprintf(`{
	"query": "query {
		   repository(name: \"%s\", owner: \"%s\") {
			discussions(%s orderBy: {field: CREATED_AT, direction: DESC}) {
				totalCount
				pageInfo {
					startCursor
					endCursor
					hasNextPage
					hasPreviousPage
				}
				edges {
					node {
						createdAt
						url
						title
						answerChosenAt
						answerChosenBy {
							login
						}
						author {
							login
						}
						answer {
							createdAt
							author {
								login
							}
							body
						}
						comments(first: %s) {
							edges {
								node {
									createdAt
									author {
										login
									}
								}
							}
						}
					}
				}
			}
		}
	}"
	}`, githubRepo, githubAccount, cursor, "100")

			// query = fmt.Sprintf(`{
			// "query": "query {
			// 	repository(name: \"%s\", owner: \"%s\") {
			// 		discussionCategories(first: 10) {
			// 				nodes {
			// 					id
			// 					name
			// 				}
			// 		}
			// 	}
			// }"
			// }`, githubRepo, githubAccount)

			log.Info(formatGraphQLStatement(query))
			currCount += 100
			// query = fmt.Sprintf(`{"query": "query { repository(name: \"nextjs\", owner: \"vercel\") { id } }"}`)
			// fmt.Println(formatGraphQLStatement(query))

			graphQLObject, err := getRepositoryData("https://api.github.com/graphql", githubToken, formatGraphQLStatement(query))
			if err != nil {
				log.Fatalf("Unable to get repository data: %+v", err)
			}
			// log.Infof("PageInfo %+v", graphQLObject.Data.Repository.Discussions.PageInfo)
			nextPage = graphQLObject.Data.Repository.Discussions.PageInfo.HasNextPage
			totalCount = graphQLObject.Data.Repository.Discussions.TotalCount
			log.Infof("totalCount: %d", totalCount)
			if nextPage {
				cursor = fmt.Sprintf("first: %s, after: \\\"%s\\\"", "100", graphQLObject.Data.Repository.Discussions.PageInfo.EndCursor)
				log.Infof("Next page: cursor = %s", cursor)
			}
			graphQLObjects = append(graphQLObjects, graphQLObject)
			// log.Info("for cursor, currCount, nextPage := initialCursor, 0, false; currCount < maxCount && (cursor == initialCursor || nextPage); {")
			// log.Infof("currCount (%d) < maxCount (%d) && (cursor (%s) == initialCursor (%s) || nextPage %v)", currCount, maxCount, cursor, initialCursor, nextPage)

		}
		// Print the GraphQL Object

		// object, _ := json.MarshalIndent(&graphQLObjects, "", "  ")
		// log.Infof("%+v\n", string(object))

		// fmt.Println("#######################################")

		for _, graphQLObject := range graphQLObjects {
			for _, discussion := range graphQLObject.Data.Repository.Discussions.Edges {
				log.Debugf("Iterating over discussion '%s'", discussion.Node.Title)
				analysis := Analysis{
					URL:         discussion.Node.URL,
					Title:       discussion.Node.Title,
					Author:      discussion.Node.Author.Login,
					CreatedAt:   discussion.Node.CreatedAt,
					AnsweredAt:  discussion.Node.Answer.CreatedAt,
					AnsweredBy:  discussion.Node.AnswerChosenBy.Login,
					ChosenAt:    discussion.Node.AnswerChosenAt,
					ChosenBy:    discussion.Node.AnswerChosenBy.Login,
					NumComments: len(discussion.Node.Comments.Edges),
				}
				count := map[string]int{}
				bestUsers := []string{}
				bestCount := 0
				for _, comment := range discussion.Node.Comments.Edges {
					user := comment.Node.Author.Login
					count[user]++
					// log.Infof("Incrementing count of %s to %d", user, count[user])
					if count[user] > bestCount {
						bestUsers = []string{user}
						bestCount = count[user]
					} else if count[user] == bestCount {
						bestUsers = append(bestUsers, user)
					}
				}
				analysis.ProlificCommentors = strings.Join(bestUsers, ",")
				analysis.ProlificCount = bestCount
				analyses = append(analyses, analysis)
			}
		}

		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		t, err := json.MarshalIndent(&analyses, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(string(t))
	} else {
		log.Infof("Loading data from %s", fileName)
		content, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(content, &analyses)
		if err != nil {
			log.Fatal(err)
		}
	}

	// sep := "|"
	// for _, analysis := range analyses {
	// 	fmt.Printf("%s\n", strings.Join([]string{
	// 		analysis.CreatedAt,
	// 		analysis.Author,
	// 		analysis.AnsweredAt,
	// 		analysis.AnsweredBy,
	// 		analysis.ChosenAt,
	// 		analysis.ChosenBy,
	// 		strconv.Itoa(analysis.NumComments),
	// 		strconv.Itoa(analysis.ProlificCount),
	// 		analysis.ProlificCommentors,
	// 	}, sep))
	// }

	type Table struct {
		DiscussionsOpenedCount int64
		DiscussionsClosedCount int64
		CloseTimes             []int64
		MedianCloseTime        int64
		MedianCloseTimeString  string
		MeanCloseTime          int64
		MeanCloseTimeString    string
	}

	type Start struct {
		Data map[int64]Table
	}

	type Interval struct {
		Data map[int64]Start
	}

	var startDate int64
	var endDate int64
	for _, analysis := range analyses {
		if startDate == 0 || epoch(analysis.CreatedAt) < startDate {
			startDate = epoch(analysis.CreatedAt)
		}
		if endDate == 0 || epoch(analysis.CreatedAt) > endDate {
			endDate = epoch(analysis.CreatedAt)
		}
	}

	log.Infof("Captured start date of %s (%d)", human(startDate), startDate)
	log.Infof("Captured end date of %s (%d)", human(endDate), endDate)

	data := make(map[int64]map[int64]Table)
	times := []int64{60 * 60 * 24, 60 * 60 * 24 * 7, 60 * 60 * 24 * 30}
	// times := []int64{1}

	for _, timeRange := range times {
		if data[timeRange] == nil {
			data[timeRange] = make(map[int64]Table)
		}
		for time := startDate; time < endDate; time = time + int64(timeRange) {
			curr := data[timeRange][time]
			for _, analysis := range analyses {
				if epoch(analysis.CreatedAt) > time &&
					epoch(analysis.CreatedAt) < time+int64(timeRange) {
					curr.DiscussionsOpenedCount++
				}
				if analysis.AnsweredAt != "" &&
					epoch(analysis.AnsweredAt) >= time &&
					epoch(analysis.AnsweredAt) < time+int64(timeRange) {
					curr.DiscussionsClosedCount++
					curr.CloseTimes = append(curr.CloseTimes, epoch(analysis.AnsweredAt)-epoch(analysis.CreatedAt))
				}
			}
			data[timeRange][time] = curr
			// log.Infof("timeRange: %d time: %d (%s)", timeRange, time, human(time))
		}
	}

	for a := range data {
		for b := range data[a] {
			curr := data[a][b]
			if len(curr.CloseTimes) == 1 {
				curr.MedianCloseTime = curr.CloseTimes[0]
				curr.MedianCloseTimeString = amountOfTime(curr.MedianCloseTime)
			} else if len(curr.CloseTimes) > 1 {
				sort.Slice(curr.CloseTimes, func(i, j int) bool { return curr.CloseTimes[i] < curr.CloseTimes[j] })
				curr.MedianCloseTime = curr.CloseTimes[len(curr.CloseTimes)/2-1]
				curr.MedianCloseTimeString = amountOfTime(curr.MedianCloseTime)

			}
			var total int64
			for _, t := range curr.CloseTimes {
				total = total + t
			}
			if curr.DiscussionsClosedCount > 0 {
				total = total / int64(curr.DiscussionsClosedCount)
			}
			curr.MeanCloseTime = total
			curr.MeanCloseTimeString = amountOfTime(total)
			data[a][b] = curr
		}
	}

	for spanIndex := range data {

		keys := make([]int64, 0, len(data[spanIndex]))
		for k := range data[spanIndex] {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		// fmt.Printf("###### %s\n", amountOfTime(spanIndex))
		// fmt.Println("date,opened,closed,meanClosed,medianClosed")

		var xValues []time.Time
		var openedValues []float64
		var totalOpenedValues []float64
		var closedValues []float64
		var totalClosedValues []float64
		var diffValues []float64
		// var totalValues []float64

		for i, k := range keys {
			c := data[spanIndex][k]
			if i == 0 {
				totalOpenedValues = append(totalOpenedValues, float64(c.DiscussionsOpenedCount))
				totalClosedValues = append(totalClosedValues, float64(c.DiscussionsClosedCount))
				diffValues = append(diffValues, float64(c.DiscussionsOpenedCount)-float64(c.DiscussionsClosedCount))
			} else {
				totalOpenedValues = append(totalOpenedValues, totalOpenedValues[i-1]+float64(c.DiscussionsOpenedCount))
				totalClosedValues = append(totalClosedValues, totalClosedValues[i-1]+float64(c.DiscussionsClosedCount))
				diffValues = append(diffValues, diffValues[i-1]+float64(c.DiscussionsOpenedCount)-float64(c.DiscussionsClosedCount))
			}

			xValues = append(xValues, time.Unix(k, 0))
			openedValues = append(openedValues, float64(c.DiscussionsOpenedCount))
			closedValues = append(closedValues, float64(c.DiscussionsClosedCount))
		}

		graph := chart.Chart{
			// XAxis: chart.XAxis{
			// 	Name: "Date",
			// },
			YAxis: chart.YAxis{
				Name: "Discussions",
			},
			Series: []chart.Series{
				chart.TimeSeries{
					Name:    "Total Opened",
					XValues: xValues,
					YValues: totalOpenedValues,
				},
				// chart.TimeSeries{
				// 	Name:    "Closed",
				// 	XValues: xValues,
				// 	YValues: closedValues,
				// },
				chart.TimeSeries{
					Name:    "Total Currently Open",
					XValues: xValues,
					YValues: diffValues,
				},
			},
		}
		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}

		fileName := strings.ReplaceAll(amountOfTime(spanIndex), " ", "")
		f, _ := os.Create(fmt.Sprintf("%s.png", fileName))
		defer f.Close()
		graph.Render(chart.PNG, f)

	}

}
