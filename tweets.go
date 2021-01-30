package main

import (
  "crypto/sha256"
  "encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/dghubble/oauth1"      // allows us to make client calls to Twitter
	"github.com/gorilla/mux"          // used for writing web handlers
)

// Tweet object allows us to consume results returned by the Twitter API
type Tweet struct {
	Date string `json:"created_at"`
	Text string `json:"text"`
	ID   string `json:"id_str"`
}

// structs returned from the oauth1 package
var config *oauth1.Config
var token *oauth1.Token

// API page limit set by Twitter (= to 3000 tweets)
const pages = 18

func getTweets(w http.ResponseWriter, r *http.Request) string {
  var maxIDQuery string // last Tweet ID of a page of results returned by the Twitter API, so we can tell the API what the next page is
	var tweets []Tweet		// placeholder slice to consume a list of Tweets
	vars := mux.Vars(r)
	userID := vars["id"]	// Twitter user handle

	// automatically authorize http.Requests
	httpClient := config.Client(oauth1.NoContext, token)

Outer:
	for i := 0; i < pages; i++ {
		/* Twitter API Request:
				- includes the user’s Twitter handle,
				- skips retweets with rts=false, and
				- puts 200 Tweets in a page (max allowed) */
		path := fmt.Sprintf("https://api.twitter.com/1.1/statuses/user_timeline.json?screen_name=%v&include_rts=false&count=200%v", userID, maxIDQuery)
		if strings.Contains(path, "favicon.ico") { // API returns this so skip it, not needed
			break
		}

		// make the HTTP request
		resp, err := httpClient.Get(path)
		if err != nil {
			respondWithError(err, w)
			break
		}

		// read returned JSON body
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			respondWithError(err, w)
			break
		}

		// unmarshal (parse JSON into Tweet struct) data into allTweets
		var allTweets []Tweet
		err = json.Unmarshal(body, &allTweets)
		if err != nil {
			respondWithError(err, w)
			break
		}

		// range through Tweets to clear out unneeded info
		for i, t := range allTweets {

			if i == len(allTweets)-1 {
				// logic to tell Twitter API where the pages are
				if maxIDQuery == fmt.Sprintf("&max_id=%v", t.ID) {
					break Outer
				}
				maxIDQuery = fmt.Sprintf("&max_id=%v", t.ID)
			}

			// remove @ mentions and links from returned Tweets
			regAt := regexp.MustCompile(`@(\S+)`)
			t.Text = regAt.ReplaceAllString(t.Text, "")
			regHTTP := regexp.MustCompile(`http(\S+)`)
			t.Text = regHTTP.ReplaceAllString(t.Text, "")
			tweets = append(tweets, t)
		}
	}

	var result []string

	for _, t := range tweets {
		result = append(result, t.Text)
	}

	stringResult := strings.Join(result, "\n")
  return stringResult
}

func handleGetTweets(w http.ResponseWriter, r *http.Request) {
	tweets := getTweets(w, r)
	w.WriteHeader(200)
	w.Write([]byte(tweets))
}

func calculateTweetHash(userID string) string {
  h := sha256.New()
  h.Write([]byte(userID))
  hashed := h.Sum(nil)
  return hex.EncodeToString(hashed)
}
