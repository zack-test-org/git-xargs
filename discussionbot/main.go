package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const defaultSleepInterval = "300"

func getChannelMap(api *slack.Client, self *slack.AuthTestResponse) (map[string]string, map[string]string, error) {
	log.Infof("Getting channel ID <=> Name mapping")
	channelMap := map[string]string{}
	reverseChannelMap := map[string]string{}
	for cursor := "init"; cursor == "init" || cursor != ""; {
		if cursor == "init" {
			cursor = ""
		}
		var channels []slack.Channel
		var err error
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: "true",
			Types:           []string{"public_channel", "private_channel"},
		}
		channels, cursor, err = api.GetConversations(params)
		if err != nil {
			return nil, nil, err
		}

		for _, channel := range channels {
			channelMap[channel.ID] = channel.Name
			reverseChannelMap[channel.Name] = channel.ID
			log.Debugf("Got channel '%s' (%s)", channel.ID, channel.Name)
		}
	}
	return channelMap, reverseChannelMap, nil
}

func getAdmins(api *slack.Client, reverseChannelMap map[string]string, userMap map[string]string, adminChannel string, self string) ([]string, error) {
	channelID := reverseChannelMap[adminChannel]
	log.Infof("Getting admins from channel '%s' (%s)", adminChannel, channelID)
	params := &slack.GetUsersInConversationParameters{
		ChannelID: channelID,
	}
	users, _, err := api.GetUsersInConversation(params)
	if err != nil {
		return nil, err
	}

	admins := []string{}
	for _, user := range users {
		if user != self {
			admins = append(admins, user)
		}
	}

	for _, admin := range admins {
		log.Infof("Admin: %s (%s)", admin, userMap[admin])
	}
	return admins, nil
}

func getUserMap(api *slack.Client) (map[string]string, map[string]string, error) {
	log.Infof("Getting user ID <=> Name mapping")
	userMap := map[string]string{}
	reverseUserMap := map[string]string{}
	users, err := api.GetUsers()
	if err != nil {
		return nil, nil, err
	}

	for _, user := range users {
		log.Debugf("Got user '%s' (%s)", user.ID, user.Name)
		userMap[user.ID] = user.Name
		reverseUserMap[user.Name] = user.ID
	}
	return userMap, reverseUserMap, nil
}

func msToTime(ms string) (time.Time, error) {
	arr := strings.Split(ms, ".")
	left, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	right, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	total := left*1000000000 + right*1000
	return time.Unix(0, total*int64(time.Nanosecond)), nil
}

func isAdmin(admins []string, user string) bool {
	for _, admin := range admins {
		if user == admin {
			return true
		}
	}
	return false
}

func alreadyReacted(api *slack.Client, channel string, timestamp string, user string, emoji string) (bool, error) {

	historyParams := &slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Latest:    timestamp,
		Inclusive: true,
		Limit:     1,
	}
	results, err := api.GetConversationHistory(historyParams)
	if err != nil {
		return true, err
	}

	for _, reaction := range results.Messages[0].Reactions {
		log.Infof("Checking %s against %s", reaction.Name, emoji)
		if reaction.Name == emoji {
			for _, u := range reaction.Users {
				log.Infof("Checking user %s against %s", u, user)
				if u == user {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}

func addReaction(api *slack.Client, channel string, timestamp string, emoji string) error {
	item := slack.ItemRef{
		Channel:   channel,
		Timestamp: timestamp,
	}
	log.Infof("Adding reaction to %s at %s\n", channel, timestamp)
	return api.AddReaction(emoji, item)
}

func removeReaction(api *slack.Client, channel string, timestamp string, emoji string) error {
	item := slack.ItemRef{
		Channel:   channel,
		Timestamp: timestamp,
	}
	log.Infof("Removing reaction to %s at %s\n", channel, timestamp)
	return api.RemoveReaction(emoji, item)
}

func getSlackThreads(api *slack.Client, channel string, timestamp string) ([]slack.Message, error) {
	log.Infof("Getting slack threads for channel %s with timestamp %s", channel, timestamp)
	repliesParameters := &slack.GetConversationRepliesParameters{
		ChannelID: channel,
		Timestamp: timestamp,
		Inclusive: true,
		Limit:     500,
	}
	r, _, _, err := api.GetConversationReplies(repliesParameters)
	return r, err
}

func getFormattedThread(api *slack.Client, channel string, timestamp string, userMap map[string]string, slackDomain string) (string, []string, error) {
	log.Infof("Creating formatted text")
	r, err := getSlackThreads(api, channel, timestamp)
	if err != nil {
		return "", nil, err
	}
	var title string
	var text []string

	for _, m := range r {
		t, err := msToTime(m.Timestamp)
		humanTime := t.Format("2006-01-02 15:04:05 UTC")
		if err != nil {
			return "", nil, err
		}

		firstLineOfText := true
		for _, line := range strings.Split(m.Text, "\n") {
			//<https://www.google.com|Google>
			re := regexp.MustCompile("<(.*?)" + regexp.QuoteMeta("|") + "(.*?)>")
			line = re.ReplaceAllString(line, "[$2]($1)")
			if title == "" {
				title = line
				ts := strings.Replace(timestamp, ".", "", -1)
				URL := fmt.Sprintf("https://%s/archives/%s/p%s", slackDomain, channel, ts)
				text = []string{fmt.Sprintf("**[%s] %s wrote in [Slack](%s)**", humanTime, userMap[m.User], URL), ""}
			} else if firstLineOfText {
				firstLineOfText = false
				text = append(text, []string{fmt.Sprintf("**[%s] %s**", humanTime, userMap[m.User]), ""}...)
			}
			if strings.HasPrefix(line, "```") {
				text = append(text, "```")
				line = strings.Replace(line, "```", "", 1)
				text = append(text, line)
			} else if strings.HasSuffix(line, "```") {
				line = strings.Replace(line, "```", "", 1)
				text = append(text, line)
				text = append(text, "```")
			} else {
				text = append(text, line)
			}
		}

	}
	return title, text, nil
}

type Author struct {
	Login string `json:"login"`
}

type Answer struct {
	Body   string `json:"body"`
	Author Author `json:"author"`
}

type Node struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Answer Answer `json:"answer"`
	Body   string `json:"body"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

type Edge struct {
	Node Node `json:"node"`
}

type DiscussionCategories struct {
	Edges []Edge `json:"edges"`
}

type Discussions struct {
	Edges []Edge `json:"edges"`
}

type Repository struct {
	ID                   string               `json:"id"`
	DiscussionCategories DiscussionCategories `json:"discussionCategories"`
	Discussions          Discussions          `json:"discussions"`
}

type Data struct {
	Repository Repository `json:"repository`
}

type TopLevel struct {
	Data Data `json:"data"`
}

/*
{
  repository(name: "discussions", owner: "pete0emerson") {
    issues(last: 10) {
      edges {
        node {
          title
          comments(last: 10) {
            edges {
              node {
                id
                body
              }
            }
          }
        }
      }
    }
  }
}

*/

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
	log.Info("JSON: ", string(t))
	return nil
}

func getRepositoryData(URL string, token string, userName string, repoName string) (TopLevel, error) {
	repo := TopLevel{}
	var jsonData = []byte(fmt.Sprintf(`{"query":"query { repository(name: \"%s\", owner: \"%s\") { id discussions(last: 10) { edges { node { id url title body answer { author { ... on User { login } } body} }}} discussionCategories(first: 10) { edges {node { id name }}}}}"}`, repoName, userName))
	//logJSON(jsonData)
	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return repo, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))
	request.Header.Set("GraphQL-Features", "discussions_api")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return repo, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	// logJSON(body)
	err = json.Unmarshal([]byte(body), &repo)
	// t, _ := json.MarshalIndent(&repo, "", "  ")
	// fmt.Println(string(t))
	return repo, err
}

func createDiscussion(title string,
	text []string,
	URL string,
	token string,
	userName string,
	repoName string) error {

	repo, err := getRepositoryData(URL, token, userName, repoName)
	if err != nil {
		log.Fatal(err)
	}
	categoryID := ""
	for _, edge := range repo.Data.Repository.DiscussionCategories.Edges {
		if edge.Node.Name == "General" {
			categoryID = edge.Node.ID
			break
		}
	}

	quotedText := strings.Join(text, "\\n")
	quotedText = strings.Replace(quotedText, "\"", "\\\\\\\"", -1)
	var jsonData = []byte(fmt.Sprintf(`{"query":"mutation { createDiscussion(input: {repositoryId: \"%s\", categoryId: \"%s\", body: \"%s\", title: \"%s\"}) { discussion { id url }}}"}`, repo.Data.Repository.ID, categoryID, quotedText, title))
	log.Infof("Sending GraphQL: %s", string(jsonData))

	request, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))
	request.Header.Set("GraphQL-Features", "discussions_api")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	logJSON(body)
	return err

}

func updateSlackThreads(repo TopLevel, api *slack.Client, self string, emoji string) error {
	// **[2021-03-02 15:07:50 UTC] pete wrote in [Slack](https://spinops.slack.com/archives/C01PVLCBCBC/p1614726470039700)**
	for _, discussion := range repo.Data.Repository.Discussions.Edges {
		if (Answer{} == discussion.Node.Answer) {
			log.Infof("No answer found for '%s'", discussion.Node.Title)
			continue
		}
		if !strings.Contains(discussion.Node.Body, "wrote in [Slack]") {
			continue
		}

		//		re := regexp.MustCompile("<(.*?)" + regexp.QuoteMeta("|") + "(.*?)>")
		re := regexp.MustCompile(regexp.QuoteMeta("wrote in [Slack](") + "(.*?)" + regexp.QuoteMeta(")"))
		match := re.FindStringSubmatch(discussion.Node.Body)
		if len(match) < 2 {
			continue
		}
		slackLink := match[1]
		// https://spinops.slack.com/archives/C01PVLCBCBC/p1614726470039700
		// fmt.Printf("%s ==> %s\n", discussion.Node.URL, slackLink)
		slackLinkArray := strings.Split(slackLink, "/")
		channel := slackLinkArray[4]
		ts := slackLinkArray[5]
		// fmt.Printf("channel: %s ts: %s\n", channel, ts)
		timestamp := strings.TrimLeft(ts, "p")
		timestamp = timestamp[:10] + "." + timestamp[10:]
		// broken INFO[0000] Getting slack threads for channel C01PVLCBCBC with timestamp 161472647.039700
		// works  INFO[0014] Getting slack threads for channel C01PVLCBCBC with timestamp 1614726470.039700
		threads, err := getSlackThreads(api, channel, timestamp)
		if err != nil {
			return err
		}
		alreadyReplied := false
		for _, m := range threads {
			if m.User == self || m.SubType == "bot_message" {
				alreadyReplied = true
				break
			}
		}
		if alreadyReplied {
			log.Infof("A bot has already answered the issue '%s'", discussion.Node.Title)
			continue
		}
		log.Infof("Replying to Slack thread %s", discussion.Node.URL)
		text := discussion.Node.Answer.Body

		re = regexp.MustCompile(regexp.QuoteMeta("[") + "(.+)" + regexp.QuoteMeta("](") + "(.+?)" + regexp.QuoteMeta(")"))
		text = re.ReplaceAllString(text, "<$2|$1>")
		text = fmt.Sprintf(`From <%s|%s> by *%s*:

%s`, discussion.Node.URL, discussion.Node.URL, discussion.Node.Answer.Author.Login, text)
		// log.Infof("%+v", text)
		_, _, err = api.PostMessage(channel,
			slack.MsgOptionIconEmoji("github"),
			slack.MsgOptionAsUser(true),
			slack.MsgOptionText(text, false),
			slack.MsgOptionTS(timestamp),
			slack.MsgOptionDisableLinkUnfurl(),
		)
		if err != nil {
			return err
		}
		addReaction(api, channel, timestamp, emoji)

	}
	return nil
}

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		log.Fatal("No SLACK_TOKEN set.")
	}

	slackDomain := os.Getenv("SLACK_DOMAIN")
	if slackDomain == "" {
		log.Fatal("No SLACK_DOMAIN set.")
	}

	adminChannel := os.Getenv("SLACK_ADMIN_CHANNEL")
	if adminChannel == "" {
		log.Fatal("No SLACK_ADMIN_CHANNEL set.")
	}

	sendEmoji := os.Getenv("SLACK_SEND_EMOJI")
	if sendEmoji == "" {
		log.Fatal("No SLACK_SEND_EMOJI set.")
	}

	lookingEmoji := os.Getenv("SLACK_LOOKING_EMOJI")
	if lookingEmoji == "" {
		log.Fatal("No SLACK_LOOKING_EMOJI set.")
	}

	answeredEmoji := os.Getenv("SLACK_ANSWERED_EMOJI")
	if answeredEmoji == "" {
		log.Fatal("No SLACK_ANSWERED_EMOJI set.")
	}

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

	api := slack.New(
		slackToken,
	)

	self, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Error getting myself: %#v\n", err)
	}

	log.Info("Updating Slack threads")
	repo, err := getRepositoryData("https://api.github.com/graphql", githubToken, githubAccount, githubRepo)
	if err != nil {
		log.Warn("Unable to get repository data")
	} else {
		err := updateSlackThreads(repo, api, self.UserID, answeredEmoji)
		if err != nil {
			log.Warn("Unable to update Slack threads")
		}
	}

	sleepIntervalInt, err := strconv.Atoi(sleepInterval)
	if err != nil {
		log.Warnf("Unable to convert '%s' to an integer, using '%s' instead", sleepInterval, defaultSleepInterval)
		sleepIntervalInt, err = strconv.Atoi(sleepInterval)
		if err != nil {
			log.Fatal("Unable to generate a good sleep interval")
		}

	}
	log.Infof("Checking for updates every %d seconds", sleepIntervalInt)
	ticker := time.NewTicker(time.Duration(sleepIntervalInt) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info("Updating Slack threads")
				repo, err := getRepositoryData("https://api.github.com/graphql", githubToken, githubAccount, githubRepo)
				if err != nil {
					log.Warn("Unable to get repository data")
				} else {
					err := updateSlackThreads(repo, api, self.UserID, answeredEmoji)
					if err != nil {
						log.Warn("Unable to update Slack threads")
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	channelMap, reverseChannelMap, err := getChannelMap(api, self)
	if err != nil {
		log.Fatalf("Unable to get channel map: %#v\n", err)
	}

	// userMap, reverseUserMap, err := getUserMap(api)
	userMap, _, err := getUserMap(api)
	if err != nil {
		log.Fatalf("Unable to get user map: %#v\n", err)
	}

	admins, err := getAdmins(api, reverseChannelMap, userMap, adminChannel, self.UserID)

	// Join this channel
	// Get channel ID from name
	// GetMembers

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		// case *slack.MessageEvent:
		// 	msg := ev.Msg
		// 	log.Infof("Got a message in %s (%s): %s", msg.Channel, channelMap[msg.Channel], msg.Text)
		case *slack.ConnectedEvent:
			log.Info("Connected to Slack")

		case *slack.InvalidAuthEvent:
			log.Fatal("Invalid token")
			return
		case *slack.ReactionAddedEvent:
			if ev.Reaction != sendEmoji {
				continue
			}
			if !isAdmin(admins, ev.User) {
				continue
			}

			log.Infof("%s added a reaction of %s in %s", userMap[ev.User], ev.Reaction, channelMap[ev.Item.Channel])

			reacted, err := alreadyReacted(api, ev.Item.Channel, ev.Item.Timestamp, self.UserID, lookingEmoji)
			if err != nil {
				log.Fatal(err)
			}

			if reacted {
				log.Infof("Bot already has reacted to this message")
				// Remove these next four lines once you're ready for the reactions to be sticky.
				err := removeReaction(api, ev.Item.Channel, ev.Item.Timestamp, lookingEmoji)
				if err != nil {
					log.Warnf("Error removing reaction %s: %+v", lookingEmoji, err)
				}
				err = removeReaction(api, ev.Item.Channel, ev.Item.Timestamp, answeredEmoji)
				if err != nil {
					log.Warnf("Error removing reaction %s: %+v", answeredEmoji, err)
				}

				continue
			}

			title, text, err := getFormattedThread(api, ev.Item.Channel, ev.Item.Timestamp, userMap, slackDomain)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(title)
			fmt.Println(strings.Join(text, "\n"))

			createDiscussion(title,
				text,
				"https://api.github.com/graphql",
				githubToken,
				githubAccount,
				githubRepo,
			)

			// fmt.Printf("\n### %s\n%s\n", title, strings.Join(text, "\n"))
			err = addReaction(api, ev.Item.Channel, ev.Item.Timestamp, lookingEmoji)
			if err != nil {
				log.Warnf("Error adding reaction: %+v", err)
			}

			// text := fmt.Sprintf(

			// _, _, err = api.PostMessage(ev.Item.Channel,
			// 	slack.MsgOptionIconEmoji("github"),
			// 	slack.MsgOptionAsUser(true),
			// 	slack.MsgOptionText(text, false),
			// 	slack.MsgOptionTS(ev.Item.Timestamp),
			// 	slack.MsgOptionDisableLinkUnfurl(),
			// )
			// if err != nil {
			// 	log.Warnf("Unable to add comment: %+v", err)
			// }

		}

	}
}
