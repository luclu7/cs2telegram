package main

import (
	"github.com/mmcdole/gofeed"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"log"
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

func parseflux(url string) (feed *gofeed.Feed) {
	parser := gofeed.NewParser()
	feed, _ = parser.ParseURL(url)
	return feed
}

func read(file string) string {
	buf, _ := ioutil.ReadFile(file)
	final := string(buf)
	return final
}

func checkandpost(b *tb.Bot) {
	log.Print("Downloading and parsing the RSS feed...")
	feed := parseflux("https://www.commitstrip.com/fr/feed/")

	comic := read("latestcomic.txt")

	f, err := os.Create("latestcomic.txt")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	var toChannel tb.Recipient = RecipientType{Channel: "@commitstrip_fr"}
	if comic == feed.Items[0].GUID {
		log.Print("No new comic available")
		f.WriteString(feed.Items[0].GUID)
	} else {
		log.Print("New comic available")
		r, _ := regexp.Compile(`<img[^>]+src="([^">]+)"`)
		matches := r.FindStringSubmatch(feed.Items[0].Content)
		f.WriteString(feed.Items[0].GUID)
		picture := &tb.Photo{File: tb.FromURL(matches[1])}
		message := feed.Items[0].Title + " " + feed.Items[0].Link
		b.Send(toChannel, message)
		b.Send(toChannel, picture)

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
