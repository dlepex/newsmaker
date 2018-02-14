package main

import (
	"time"

	"github.com/dlepex/newsmaker/news"
)

type Config struct {
	RTick     duration            `toml:"rotation_tick"`
	MuteHours *[2]int             `toml:"mute_hours"`
	Filters   []*FilterConf       `toml:"filters"`
	Sources   map[string]*SrcConf `toml:"src"`
	Pubs      map[string]*PubConf `toml:"pub"`
}

type FilterConf struct {
	Cond    string   `toml:"cond"`
	Sources []string `toml:"sources"`
	Pubs    []string `toml:"pubs"`
}

type SrcConf struct {
	CD    duration `toml:"cd"`
	Links []string `toml:"links"`
	Categ []string `toml:"categ"`
}

type PubConf struct {
	SendPause duration `toml:"send_pause"`
	GetUrl    string   `toml:"get_url"`
}

func NewPipeline(c *Config) (pl news.Pipeline, ers []error) {
	pl = news.NewPipeline()
	globalMuteHours := news.DayInterval{}
	if c.MuteHours != nil {
		globalMuteHours = news.DayHoursFromTo(c.MuteHours[0], c.MuteHours[1])
	}
	check := func(e error) bool {
		if e == nil {
			return true
		}
		ers = append(ers, e)
		return false
	}
	for n, c := range c.Pubs {
		pub, err := c.toPub(n)
		if check(err) {
			check(pl.AddPublisher(pub))
		}
	}
	for n, c := range c.Sources {
		src, err := c.toSource(n, globalMuteHours)
		if check(err) {
			check(pl.AddSource(src))
		}
	}
	for _, c := range c.Filters {
		check(pl.AddFilter(c.toFilter()))
	}
	if len(ers) != 0 {
		pl = nil
	}

	return
}

func (c *FilterConf) toFilter() *news.Filter {
	return &news.Filter{Cond: c.Cond, Sources: c.Sources, Pubs: c.Pubs}
}

func (c *SrcConf) toSource(n string, muteHours news.DayInterval) (news.Source, error) {
	return news.NewFeedSrc(news.FeedSrcParams{
		SourceInfo: news.SourceInfo{
			Name:         n,
			Categories:   c.Categ,
			Cooldown:     c.CD.Duration,
			MuteInterval: muteHours,
		},
		Links: c.Links,
	})
}

func (c *PubConf) toPub(n string) (news.Pub, error) {
	return &news.URLPub{
		PubInfo: news.PubInfo{
			Name: n,
		},
		Link:  c.GetUrl,
		Pause: c.SendPause.Duration,
	}, nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
