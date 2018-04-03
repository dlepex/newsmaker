package news

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"text/template"

	"go.uber.org/zap"
)

//
//  It's a playground, not test-cases.
//
func TestTelegram(t *testing.T) {
	msg := `
		01.01.2017
		*Title News Happened*
		[dfdf](http://www.dell.com/learn/ru/ru/rudhs1/campaigns/faq-ru)
	`

	tpl := "https://api.telegram.org/bot500336234:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown"

	tpl = "https://api.telegram.org/bot500336234:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001356941917&parse_mode=Markdown"
	link := fmt.Sprintf(tpl, url.QueryEscape(msg))
	resp, err := http.Get(link)
	slog.Info(resp, err)
}

// "https://api.telegram.org//sendMessage?text=Hello&chat_id=-1001159128561"
//  -1001356941917
//  // "https://news.rambler.ru/rss/economics", // "http://www.aif.ru/rss/money.php", //TdPO //"http://www.aif.ru/rss/money.php", //"https://lenta.ru/rss/news", // //"http://www.argus.ru/news/?type=rss", // "https://news.rambler.ru/rss/economics/",
func TestFeedSourc(t *testing.T) {
	SetLoggerVerbose()
	src, err := NewFeedSrc(FeedSrcParams{
		SourceInfo: SourceInfo{
			Name: "regnum",
		},
		Links: []string{"https://news.yandex.ru/business.rss", "https://news.yandex.ru/politics.rss", "https://news.yandex.ru/finances.rss"},
	})

	if err != nil {
		panic(err)
	}

	FeedSrcDebug = true
	FeedSrcPauseRand = 5 * time.Second
	FeedSrcPause = 5 * time.Second

	items := []*Item{}

	sink := func(it *Item) {
		items = append(items, it)
	}

	src.Receive(sink)

	for _, it := range items {
		slog.Info(it.Link)
	}
	slog.Info(len(items))
}

func TestRotator(t *testing.T) {

	logger := zap.NewExample()
	SetLogger(logger)
	rot := rotator{
		Tick: time.Second * 1,
		Elems: []rotatorElem{
			rotatorElem{
				Cooldown: time.Second * 2,
				Fn: func(now time.Time) {
					slog.Info("2")
				},
			}, /*
				RotatorElem{
					Cooldown: time.Second * 3,
					Fn: func() {
						slog.Info("3")
					},
				},
				RotatorElem{
					Cooldown: time.Second * 5,
					Fn: func() {
						slog.Info("5")
					},
				},*/
		},
	}
	quit := make(chan struct{})

	go rot.run(quit)

	time.Sleep(7 * time.Second)
	quit <- struct{}{}
	time.Sleep(15 * time.Second)
}

func TestMute(t *testing.T) {
	di := DayInterval{}
	now := time.Now()
	di = DayHoursFromTo(20, 4)
	fmt.Printf("!!!%v %s!!!!!!", di.ContainsTime(now), di)
}

func TestURLPub(t *testing.T) {
	pub := NewHTTPPub(&HTTPPubParams{
		PubInfo: PubInfo{
			Name: "test",
		},
		Link:  "https://api.telegram.org/bot500336234:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown",
		Pause: 7 * time.Second,
	})

	ch := make(chan *Item, 20)

	go pub.Publish(ch)

	src := &SourceInfo{
		Name: "testSrc",
	}
	now := time.Now()

	ch <- &Item{
		ItemParams: ItemParams{
			Src:       src,
			Title:     "Title test#1",
			Link:      "https://gobyexample.com/time-formatting-parsing",
			Published: &now,
		},
	}

	ch <- &Item{
		ItemParams: ItemParams{
			Src:   src,
			Title: "Title test#2",
			Link:  "https://gobyexample.com/random-numbers",
		},
	}

	time.Sleep(30 * time.Second)

}
func TestPipeline(t *testing.T) {
	SetLoggerVerbose()
	src1, err := NewFeedSrc(FeedSrcParams{
		SourceInfo: SourceInfo{
			Name: "regnum",
		},
		Links: []string{"https://regnum.ru/rss/polit"},
	})

	if err != nil {
		panic(err)
	}

	f1 := &Filter{
		Cond:    "россия",
		Sources: []string{"reg"},
		Pubs:    []string{"p1", "test"},
	}

	f2 := &Filter{
		Cond:    "москва",
		Sources: []string{"reg"},
		Pubs:    []string{"p2"},
	}

	check := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	pub := NewHTTPPub(&HTTPPubParams{
		PubInfo: PubInfo{
			Name: "test",
		},
		Link:  "https://api.telegram.org/bot500336234:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown",
		Pause: 3 * time.Second,
	})

	p1, err := NewLogPub(PubInfo{
		Name: "p1",
	})
	check(err)

	p2, err := NewLogPub(PubInfo{
		Name: "p2",
	})
	check(err)
	pl := NewPipelineDefault()

	err = pl.AddSource(src1)
	check(err)
	err = pl.AddFilter(f1)
	check(err)
	err = pl.AddFilter(f2)
	check(err)
	err = pl.AddPublisher(p1)
	check(err)
	err = pl.AddPublisher(p2)
	check(err)
	err = pl.AddPublisher(pub)
	check(err)

	check(pl.Run())

	slog.Debug(pl)

	pl.Wait()

}

func TestTpl(t *testing.T) {
	tmpl, _ := template.New("test").Parse("*{{.Title}}* {{.DateFmt}} \n{{.Src.Name}} {{.Link}}")

	item := &Item{ItemParams: ItemParams{
		Src: &SourceInfo{
			Name: "Hello",
		},
		Link:  "link/aa/aaa/a",
		Title: "bbb",
	}, DateFmt: "12.121.1"}
	buf := bytes.NewBuffer(make([]byte, 0, 256))

	tmpl.Execute(buf, item)
	t.Log(buf.String())
}
