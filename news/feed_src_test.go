package news

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestTelegram(t *testing.T) {

	msg := `
		01.01.2017
		*Title News Happened*
		[finam](http://www.dell.com/learn/ru/ru/rudhs1/campaigns/faq-ru)
	`

	tpl := "https://api.telegram.org/bot500397448:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001159128561&parse_mode=Markdown"

	tpl = "https://api.telegram.org/bot500397448:AAGuVfL-EP-Wnlj1v5Ub40oykOkxrJgZZbI/sendMessage?text=%s&chat_id=-1001356941917&parse_mode=Markdown"
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
	rot := Rotator{
		Tick: time.Second * 1,
		Elems: []RotatorElem{
			RotatorElem{
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

	go rot.Run(quit)

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
