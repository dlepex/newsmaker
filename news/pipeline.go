package news

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Usually there's 1 instance per app.
type Pipeline = *PipelineOpaq

type PipelineOpaq struct {
	sources map[string]*srcData
	pubs    map[string]*pubData
	filters []*Filter
	modLock sync.Mutex // guards modification and start
	started bool
	wg      sync.WaitGroup
	dedup   SyncDeduplicator
	rot     Rotator
	quit    chan struct{}
	prodc   chan *Item
}

type srcData struct {
	Source
	quit      chan struct{}
	filterInd []int
	guard     Guard
}

type pubData struct {
	Pub
	ch chan *Item
}

func NewPipeline() Pipeline {
	return &PipelineOpaq{
		dedup:   SyncDeduplicator{Deduplicator: NewDedup(8 * 1024)},
		sources: make(map[string]*srcData),
		pubs:    make(map[string]*pubData),
	}
}

func (pl Pipeline) modify(modFn func() error) error {
	pl.modLock.Lock()
	defer pl.modLock.Unlock()
	if pl.started {
		return errors.New("Pipeline can be modified only before it is started!")
	}
	return modFn()
}

func (pl Pipeline) AddSource(s Source) error {
	return pl.modify(func() error {
		n := s.Info().Name
		if _, has := pl.sources[n]; has {
			return fmt.Errorf("duplicate source: %s", n)
		}
		pl.sources[n] = &srcData{Source: s}
		return nil
	})
}

func (pl Pipeline) AddPublisher(p Pub) error {
	return pl.modify(func() error {
		n := p.Info().Name
		if _, has := pl.pubs[n]; has {
			return fmt.Errorf("duplicate publisher: %s", n)
		}
		pl.pubs[n] = &pubData{Pub: p}
		return nil
	})
}

func (pl Pipeline) AddFilter(f *Filter) error {
	return pl.modify(func() error {
		if err := f.init(); err != nil {
			return err
		}
		pl.filters = append(pl.filters, f)
		return nil
	})
}

func (pl Pipeline) beforeStart() error {
	pl.modLock.Lock()
	defer pl.modLock.Unlock()
	if pl.started {
		return errors.New("pl: already started")
	}
	if len(pl.filters) == 0 {
		return errors.New("pl: no filters")
	}

	for _, s := range pl.sources {
		info := s.Info()
		s.filterInd = chooseFilters(pl.filters, func(f *Filter) bool {
			return f.matchSrc(info)
		})
		if len(s.filterInd) == 0 {
			slog.Infow("pl: orphaned source", "name", info.Name)
			delete(pl.sources, info.Name)
		}
	}
	if len(pl.sources) == 0 {
		return errors.New("No sources")
	}
	for _, p := range pl.pubs {
		info := p.Info()
		indices := chooseFilters(pl.filters, func(f *Filter) bool {
			return f.matchPub(info)
		})
		if len(indices) == 0 {
			slog.Infow("pl: remove pub", "name", info.Name)
			delete(pl.pubs, info.Name)
		} else {
			for _, i := range indices {
				f := pl.filters[i]
				f.pubs = append(f.pubs, info.Name)
			}
		}
	}
	if len(pl.pubs) == 0 {
		return errors.New("No pubs")
	}
	pl.started = true
	return nil
}

func (pl Pipeline) Start() error {
	const chanSize = 2 * 1024
	if err := pl.beforeStart(); err != nil {
		return err
	}
	pl.prodc = make(chan *Item, 2*chanSize)
	pl.wg.Add(len(pl.pubs))
	for _, p := range pl.pubs {
		p.ch = make(chan *Item, chanSize)
		pub := p
		go func() {
			defer pl.wg.Done()
			pub.Publish(pub.ch)
		}()
	}

	for _, _s := range pl.sources {
		s, info := _s, _s.Info()
		sink := info.newSink(pl.prodc)
		pl.rot.Elems = append(pl.rot.Elems, RotatorElem{
			Cooldown: info.Cooldown,
			Fn: func(now time.Time) {
				if info.MuteInterval.ContainsTime(now) {
					return
				}
				s.guard.Go(func() {
					if err := s.Receive(sink); err != nil {
						log.WithField("src", info.Name).Error(err)
					}
				})
			},
		})
	}
	pl.quit = make(chan struct{})
	GoWG(&pl.wg, func() {
		pl.rot.Run(pl.quit)
	})
	GoWG(&pl.wg, pl.run)
	return nil
}

func (pl Pipeline) run() {

	pubs := make(map[string]struct{})

	for it := range pl.prodc {
		_, ok := pl.sources[it.Src.Name]
		if !ok {
			log.Fatal("item source not found (bug!)", it.Src.Name)
			continue
		}

		atLeastOneMatch := false
		for _, f := range pl.filters {
			if f.dnf.MatchWords(it.words) {
				atLeastOneMatch = true
				for _, pname := range f.pubs {
					pubs[pname] = struct{}{}
				}
			}
		}

		if !atLeastOneMatch {
			continue
		}

		if pl.dedup.Check(it.key) {
			continue
		}

		for pname, _ := range pubs {
			logEvent := "pub_send"
			select {
			case pl.pubs[pname].ch <- it:
			default:
				logEvent = "pub_full"
			}
			slog.Infow(logEvent, "pub", pname, "title", it.Title, "link", it.Link, "src", it.Src.Name, "key", it.key)
			delete(pubs, pname)
		}
	}
}

func (pl Pipeline) Stop() {
	pl.quit <- struct{}{}
	close(pl.prodc)
	for _, p := range pl.pubs {
		close(p.ch)
	}
}

func (pl Pipeline) Wait() {
	pl.wg.Wait()
}
