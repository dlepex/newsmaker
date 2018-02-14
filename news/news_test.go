package news

import (
	"testing"
	"time"
)

func TestURLPub(t *testing.T) {
	pub := &URLPub{
		PubInfo: PubInfo{
			Name: "test",
		},
		Link:  "https://api.telegram.org/bot500397448:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown",
		Pause: 7 * time.Second,
	}

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
	RotatorTickDefault = 10 * time.Second
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
		Cond:    "мид;медведев",
		Sources: []string{"reg"},
		Pubs:    []string{"p1", "test"},
	}

	f2 := &Filter{
		Cond:    "украин,донбас;путин",
		Sources: []string{"reg"},
		Pubs:    []string{"p2"},
	}

	check := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	pub := &URLPub{
		PubInfo: PubInfo{
			Name: "test",
		},
		Link:  "https://api.telegram.org/bot500397448:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown",
		Pause: 3 * time.Second,
	}

	p1, err := NewLogPub(PubInfo{
		Name: "p1",
	})
	check(err)

	p2, err := NewLogPub(PubInfo{
		Name: "p2",
	})
	check(err)
	pl := NewPipeline()

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

	check(pl.Start())

	slog.Debug(pl)

	pl.Wait()

}
