package news

import (
	"errors"
	"net/http"
	"time"

	"github.com/dlepex/newsmaker/words"

	"github.com/dlepex/newsmaker/strext"
)

// SourceInfo contains description of the news source (producer)
type SourceInfo struct {
	// Id uniquely identifies news source, and by convention must be have the following format
	// AgencyName.ChannelName
	Name string
	// This field will be calculated based on Id field, if not set
	Agency string
	// Categories is used to "prefilter" content based on categories/tags matching
	Categories   []string
	Cooldown     time.Duration
	MuteInterval DayInterval
}

type SharedCtx struct {
	HTTPClient *http.Client
}

type Source interface {
	Info() *SourceInfo
	Receive(sink func(*Item)) error
}

type ItemParams struct {
	Src        *SourceInfo
	Title      string
	Link       string
	Categories []string
	Published  *time.Time
}

// Item is "the news item" produced by Source
// Item must be properly initialized by calling Init() function
type Item struct {
	ItemParams
	words []string // title words
	key   DedupKey
}

type PubInfo struct {
	Name string
}

type Pub interface {
	Info() *PubInfo
	Publish(in <-chan *Item)
}

func (s *SourceInfo) Check() error {
	if strext.IsBlank(s.Name) {
		return errors.New("source: name required")
	}
	if s.Cooldown == 0 {
		s.Cooldown = 15 * time.Minute
	}
	return nil
}

func NewItem(p ItemParams) (*Item, error) {
	if strext.IsBlank(p.Title) {
		return nil, errors.New("title required")
	}
	it := &Item{ItemParams: p}
	it.words = words.Split(it.Title)
	it.key = StrToDedupKey(it.words...)
	return it, nil
}

func (src *SourceInfo) newSink(ch chan<- *Item) func(*Item) {

	if len(src.Categories) == 0 {
		return func(it *Item) {
			ch <- it
		}
	}

	return func(it *Item) {
		if !matchAnyGlobAny(it.Categories, src.Categories) {
			return
		}
		ch <- it
	}
}

type logPub struct {
	PubInfo
}

func NewLogPub(p PubInfo) (Pub, error) {
	return &logPub{p}, nil
}

func (pub *logPub) Info() *PubInfo {
	return &pub.PubInfo
}

func (pub *logPub) Publish(ch <-chan *Item) {
	for it := range ch {
		slog.Infow("log_pub_send", "pub", pub.Name, "title", it.Title, "link", it.Link, "src", it.Src.Name, "at", it.Published)
	}
}

var URLPubPause time.Duration = time.Second

type URLPub struct {
	PubInfo
	Link   string
	Pause  time.Duration
	Client *http.Client
}

func (pub *URLPub) Info() *PubInfo {
	return &pub.PubInfo
}

func (info *PubInfo) PublishByOne(ch <-chan *Item, delay time.Duration, fn func(*Item) error) {

	for it := range ch {
		if err := fn(it); err != nil {
			slog.Infow("pub_error", "pub", info.Name, "err", err, "key", it.key)
		}
		time.Sleep(delay)
	}
}

func (pub *URLPub) Publish(ch <-chan *Item) {

	/*
		for it := range ch {
			timeStr := ""
			if it.Published != nil {
				timeStr = it.Published.Format("02.01 15:04")
			}

			msg := fmt.Sprintf("*%s*  %s\n [%s](%s)", it.Title, timeStr, it.Src.Name, it.Link)
			link := fmt.Sprintf(pub.Link, url.QueryEscape(msg))

			r, err := cl.Get(link)

			cl.Do()
			var st string
			if r != nil {
				st = r.Status
			}
			defer r.Body.Close()
			time.Sleep(pause)
		}*/
}
