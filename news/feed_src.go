package news

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	gfd "github.com/mmcdole/gofeed"
)

//FeedSrcParams -
type FeedSrcParams struct {
	SourceInfo
	Links  []string
	Client *http.Client
}

type feedSrc struct {
	FeedSrcParams
	links []string
}

// FeedSrcDebug - log extra information
var FeedSrcDebug bool

var (
	// FeedSrcPause - multi-link feedsrc: pause between each link request
	FeedSrcPause = 20 * time.Second
	// FeedSrcPauseRand - feedsrc: random pause between each link request
	FeedSrcPauseRand = 20 * time.Second
)

//NewFeedSrc creates feed (rss/atom) source
func NewFeedSrc(p FeedSrcParams) (Source, error) {
	if p.Client == nil {
		p.Client = http.DefaultClient
	}
	if len(p.Links) == 0 {
		return nil, errors.New("feed src: no links")
	}
	if err := p.Check(); err != nil {
		return nil, err
	}
	slog.Debugw("created source", "src", p.Name, "cd", p.Cooldown, "links", p.Links, "mute-hours", p.MuteInterval)
	return &feedSrc{p, append([]string{}, p.Links...)}, nil
}

func (src *feedSrc) Info() *SourceInfo {
	return &src.SourceInfo
}

//shuffleLinks uses Sattolo's algorithm
func (src *feedSrc) shuffleLinks() []string {
	links := src.links
	for i := len(links); i > 1; {
		i--
		j := rand.Intn(i)
		links[i], links[j] = links[j], links[i]
	}
	return links
}

func (src *feedSrc) Receive(sink func(*Item)) error {
	for _, link := range src.shuffleLinks() {
		src.ReceiveOne(link, sink)
		time.Sleep(FeedSrcPause + time.Duration(rand.Int63n(int64(FeedSrcPauseRand))))
	}
	return nil
}

func (src *feedSrc) ReceiveOne(link string, sink func(*Item)) {
	p := gfd.NewParser()
	feed, err := p.ParseURL(link)
	if err != nil {
		slog.Errorw(err.Error(), "src", src.Name, "link", link)
		return
	}
	feedItems := feed.Items
	feed.Items = nil
	src.debug("feed", feed)
	slog.Debugw("feed_receive", "link", link, "count", len(feedItems))
	for _, v := range feedItems {
		v.Description = ""
		src.debug("item", v)

		params := ItemParams{
			Link:      v.Link,
			Title:     v.Title,
			Published: v.PublishedParsed,
			Src:       &src.SourceInfo,
		}

		item, err := NewItem(params)
		if err != nil {
			slog.Errorw("feed_parse_error", "err", err, "link", v.Link, "params", params)
			continue
		}
		sink(item)
	}
}

func (src *feedSrc) debug(what string, value interface{}) {
	if FeedSrcDebug {
		slog.Debugw("feed_debug", "src", src.Name, "what", what, "value", value)
	}
}
