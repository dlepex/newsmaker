package news

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
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

// Source - news source interface
// Receive is called if this source is "elected" by rotation, all received items should be send to sink() function.
type Source interface {
	Info() *SourceInfo
	Receive(sink func(*Item)) error
}

// ItemParams - various news item params received from feed.
type ItemParams struct {
	Src        *SourceInfo
	Title      string
	Link       string
	Categories []string
	Published  *time.Time
}

// Item is "the news item" produced by Source
type Item struct {
	ItemParams
	words   []string // title words
	key     DedupKey
	DateFmt string // formated datetime (for text template use only)
}

// PubInfo - publisher description
type PubInfo struct {
	Name string
}

// Pub aka publisher/notifier.
// Publish: message receive loop, if channel `in` is closed - the loop should exit.
type Pub interface {
	Info() *PubInfo
	Publish(in <-chan *Item)
}

// Check - verifies source info consistency
func (s *SourceInfo) Check() error {
	if strext.IsBlank(s.Name) {
		return errors.New("source: name required")
	}
	if s.Cooldown == 0 {
		s.Cooldown = 15 * time.Minute
	}
	return nil
}

// NewItem - item constructor from params
func NewItem(p ItemParams) (*Item, error) {
	if strext.IsBlank(p.Title) {
		return nil, errors.New("title required")
	}
	it := &Item{ItemParams: p}
	it.words = words.Split(it.Title)
	it.key = StrToDedupKey(it.words...)
	return it, nil
}

func (s *SourceInfo) newSink(ch chan<- *Item) func(*Item) {
	if len(s.Categories) == 0 {
		return func(it *Item) {
			ch <- it
		}
	}
	return func(it *Item) {
		if !matchAnyGlobAny(it.Categories, s.Categories) {
			return
		}
		ch <- it
	}
}

// dummy publisher that only logs items.
type logPub struct {
	PubInfo
}

func NewLogPub(p PubInfo) (Pub, error) { //nolint
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

var URLPubPause time.Duration = time.Second //nolint:golint

type ItemStringer func(it *Item) string //nolint:golint

type HTTPPub struct { //nolint:golint
	HTTPPubParams
	client *http.Client
}
type HTTPPubParams struct { //nolint:golint
	PubInfo
	Link         string        // url sprinf format, the only argument of sprintf is item string.
	ItemStringer               // converts item to string
	Pause        time.Duration // pause between http queries
}

func NewHTTPPub(params *HTTPPubParams) Pub { //nolint:golint
	p := &HTTPPub{
		HTTPPubParams: *params,
		client:        &http.Client{},
	}
	if p.ItemStringer == nil {
		p.ItemStringer = itemStrDefault
	}
	return p
}

func itemStrDefault(it *Item) string {
	return fmt.Sprintf("%v", it)
}

func (pub *HTTPPub) Info() *PubInfo { //nolint:golint
	return &pub.PubInfo
}

func (info *PubInfo) PublishByOne(ch <-chan *Item, delay time.Duration, publish func(*Item) error) { //nolint:golint

	for it := range ch {
		if err := publish(it); err != nil {
			slog.Infow("pub_error", "pub", info.Name, "err", err, "key", it.key)
		}
		time.Sleep(delay)
	}
}

func (pub *HTTPPub) Publish(ch <-chan *Item) { //nolint:golint
	pub.PublishByOne(ch, pub.Pause, func(it *Item) error {
		if it.Published != nil {
			it.DateFmt = it.Published.Format("02.01 15:04")
		}
		msg := pub.ItemStringer(it)
		link := fmt.Sprintf(pub.Link, url.QueryEscape(msg))
		r, err := pub.client.Get(link)
		if err != nil {
			return err
		}
		defer r.Body.Close() // nolint:errcheck
		if r != nil {
			st := r.StatusCode
			if !(200 <= st && st < 300) {
				return fmt.Errorf("bad http status: %v (%s)", st, r.Status)
			}
		}
		return nil
	})
}

func NewItemTemplateStringer(gotmpl string) ItemStringer { //nolint:golint
	t := template.Must(template.New("item-template").Parse(gotmpl))
	return func(it *Item) string {
		buf := bytes.NewBuffer(make([]byte, 0, 256))
		t.Execute(buf, it) //nolint:errcheck
		return buf.String()
	}
}
