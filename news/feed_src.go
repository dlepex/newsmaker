package news

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	gfd "github.com/mmcdole/gofeed"
)

type FeedSrcParams struct {
	SourceInfo
	Links  []string
	Client *http.Client
}

type feedSrc struct {
	FeedSrcParams
	links []string
}

var FeedSrcDebug bool
var FeedSrcPause time.Duration = 20 * time.Second
var FeedSrcPauseRand time.Duration = 20 * time.Second

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

func (src *feedSrc) Receive(sink func(*Item)) error {
	// shuffle links
	links := src.links
	for i := len(links) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		links[i], links[j] = links[j], links[i]
	}
	for _, link := range links {
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

func (src *feedSrc) Sleep(dur int64, quit <-chan struct{}) bool {
	select {
	case <-time.After(time.Duration(dur)):
		return true
	case <-quit:
		return false
	}
}

func (src *feedSrc) debug(what string, value interface{}) {
	if FeedSrcDebug {
		slog.Debugw("feed_debug", "src", src.Name, "what", what, "value", value)
	}
}