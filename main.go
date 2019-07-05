package main

import (
	"github.com/mmcdole/gofeed"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// RecipientType - RecipientType
type RecipientType struct {
	Channel string `json:"channel"`
}

// Recipient - Recipient
func (x RecipientType) Recipient() string {
	return x.Channel
}

func parseflux(url string) (feed *gofeed.Feed, err error) {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{Timeout: time.Second * 10}
	feed, err = parser.ParseURL(url)
	return feed, err
}

func read(file string) string {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	final := string(buf)
	return final
}

func checkandpost(b *tb.Bot) {
	log.Print("Downloading and parsing the RSS feed...")
	feed, err := parseflux("https://www.commitstrip.com/fr/feed/")
	if err == nil {
		comic := read("latestcomic.txt")
		f, err := os.Create("latestcomic.txt")
		if err != nil {
			log.Fatal(err)
			return
		}

		defer f.Close()

		var toChannel tb.Recipient = RecipientType{Channel: "@commitstrip_fr"}
		if comic == feed.Items[0].GUID {
			log.Print("No new comic available")
			f.WriteString(feed.Items[0].GUID)
		} else {
			log.Print("New comic available")
			r, err := regexp.Compile(`<img[^>]+src="([^">]+)"`)
			if err != nil {
				log.Fatal(err)
				return
			}
			matches := r.FindStringSubmatch(feed.Items[0].Content)
			f.WriteString(feed.Items[0].GUID)
			picture := &tb.Photo{File: tb.FromURL(matches[1])}
			message := feed.Items[0].Title + " " + feed.Items[0].Link
			b.Send(toChannel, message)
			b.Send(toChannel, picture)

		}
	} else {
		log.Print("Error downloading/parsing the RSS feed: " + err.Error())
	}
}

func main() {
	log.Print("Starting...")
	token := read("token.txt")
	token = strings.TrimSuffix(token, "\n")
	b, err := tb.NewBot(tb.Settings{
		Token: token,
		// You can also set custom API URL. If field is empty it equals to "https://api.telegram.org"
		//URL:    "http://195.129.111.17:8012",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	//checkandpost(b)
	dateTicker := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-dateTicker.C:
			checkandpost(b)
		}
	}

}
