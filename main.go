package main

import (
	"encoding/json"
	"fmt"
	"github.com/deckarep/gosx-notifier"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

/*

	Reddit api:
		url: reddit/r/<subreddit>.json
		resp_json[data][children] -> iterate -> i[data] -> [title], [permalink], [subreddit], [created_utc], [id], [num_comments]

			- permalink: relative path to comments; eg: /r/Python/comments/3dgkp0/consolebased_question_and_answering_system_with/
			- created: possibly unixtimestamp

*/

const (
	REDDIT_URL = "http://reddit.com"
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.134 Safari/537.36"
)

type RedditItem struct {
	Data struct {
		Title        string
		Permalink    string
		Subreddit    string
		Id           string
		Num_comments int
		Created_utc  float32
	}
}

type RedditData struct {
	Data struct {
		Children []RedditItem
	}
}

func createNotification(heading, title, subtitle, link string) error {
	//At a minimum specifiy a message to display to end-user.
	note := gosxnotifier.NewNotification(heading)
	note.Title = title
	note.Subtitle = subtitle
	note.Link = link

	//Then, push the notification
	return note.Push()

}

func readSubreddit(subreddit string) (RedditData, error) {
	url := fmt.Sprintf("%s/r/%s.json?limit=5", REDDIT_URL, subreddit)

	var data RedditData

	res, err := http.Get(url)
	if err != nil {
		return data, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func createSubredditNotification(subreddit string) {

	heading := fmt.Sprintf("Reddit - %s", subreddit)
	ticker := time.NewTicker(time.Minute)

	seen_ids := map[string]bool{}

	for {
		select {
		case <-ticker.C:
			rds, err := readSubreddit(subreddit)

			if err != nil {
				log.Fatal(err)
			}
			var body string
			var link string
			for _, child := range rds.Data.Children {
				if ok, _ := seen_ids[child.Data.Id]; ok {
					fmt.Println("Seen id: ", child.Data.Id)
					continue
				}
				fmt.Println(child.Data)
				body = fmt.Sprintf("[%d] %s", child.Data.Num_comments, child.Data.Title)
				link = fmt.Sprintf("http://reddit.com%s", child.Data.Permalink)
				createNotification(heading, child.Data.Title, body, link)

				seen_ids[child.Data.Id] = true
			}
		}
	}
}

func main() {

	subreddits := []string{
		"python",
		"golang",
		"technology",
		"programming",
	}
	ticker := time.NewTicker(time.Second * 30)

	var wg sync.WaitGroup
	wg.Add(len(subreddits))

	for _, subreddit := range subreddits {
		select {
		case <-ticker.C:
			go createSubredditNotification(subreddit)

		}
	}
	wg.Wait()
}
