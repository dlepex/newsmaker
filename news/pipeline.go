package news

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Pipeline - news filtering pipeline
// Pipeline uses rotator, that schedules source requests randomly,
// so use rand.Seed() before launching pipeline.
type Pipeline struct {
	sources map[string]*srcData
	pubs    map[string]*pubData
	filters []*Filter
	dedup   Deduplicator // deduplicator (LRU) for news titles (to avoid repeated notifications)
	rot     rotator

	chanSize int
	quit     chan struct{} // stop channel
	prodc    chan *Item    // channel to which the sources write

	lock    sync.Mutex // guards modification and launch
	started bool
	wg      sync.WaitGroup // tracks launched goroutines
}

type srcData struct {
	Source
	quit      chan struct{}
	filterInd []int
	guard     Guard // guards rotation (i.e. only 1 goroutine may read source)
}

type pubData struct {
	Pub
	ch chan *Item
}

// NewPipeline - creates pipeline
// chanSize - buffered channels size constant
func NewPipeline(chanSize int, d Deduplicator) *Pipeline {
	return &Pipeline{
		dedup:    DedupSync(d),
		sources:  make(map[string]*srcData),
		pubs:     make(map[string]*pubData),
		chanSize: chanSize,
	}
}

func NewPipelineDefault() *Pipeline { //nolint:golint
	return NewPipeline(1024, NewDedup(8192))
}

func (pl *Pipeline) modify(modFn func() error) error {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	if pl.started {
		return errors.New("pipeline can be modified only before it was started")
	}
	return modFn()
}

func (pl *Pipeline) AddSource(s Source) error { //nolint:golint
	return pl.modify(func() error {
		n := s.Info().Name
		if _, has := pl.sources[n]; has {
			return fmt.Errorf("duplicate source: %s", n)
		}
		pl.sources[n] = &srcData{Source: s}
		return nil
	})
}

func (pl *Pipeline) AddPublisher(p Pub) error { //nolint:golint
	return pl.modify(func() error {
		n := p.Info().Name
		if _, has := pl.pubs[n]; has {
			return fmt.Errorf("duplicate publisher: %s", n)
		}
		pl.pubs[n] = &pubData{Pub: p}
		return nil
	})
}

func (pl *Pipeline) AddFilter(f *Filter) error { //nolint:golint
	return pl.modify(func() error {
		if err := f.init(); err != nil {
			return err
		}
		pl.filters = append(pl.filters, f)
		return nil
	})
}

func (pl *Pipeline) beforeStart() error {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	if pl.started {
		return errors.New("pipeline: already started")
	}
	if len(pl.filters) == 0 {
		return errors.New("pipeline: no filters")
	}

	for _, s := range pl.sources {
		info := s.Info()
		s.filterInd = chooseFilters(pl.filters, func(f *Filter) bool {
			return f.matchSrc(info)
		})
		if len(s.filterInd) == 0 {
			slog.Infow("pipeline: orphaned source", "name", info.Name)
			delete(pl.sources, info.Name)
		}
	}
	if len(pl.sources) == 0 {
		return errors.New("pipeline: no sources")
	}
	for _, p := range pl.pubs {
		info := p.Info()
		indices := chooseFilters(pl.filters, func(f *Filter) bool {
			return f.matchPub(info)
		})
		if len(indices) == 0 {
			slog.Infow("pipeline: remove pub", "name", info.Name)
			delete(pl.pubs, info.Name)
		} else {
			for _, i := range indices {
				f := pl.filters[i]
				f.pubs = append(f.pubs, info.Name)
			}
		}
	}
	if len(pl.pubs) == 0 {
		return errors.New("pipeline: no publishers(aka notifiers)")
	}
	pl.started = true
	return nil
}

// Run - launches the pipeline.
func (pl *Pipeline) Run() error {
	if err := pl.beforeStart(); err != nil {
		return err
	}
	// create producer channel: channel to which the sources write
	pl.prodc = make(chan *Item, 2*pl.chanSize)
	pl.wg.Add(len(pl.pubs))
	for _, p := range pl.pubs {
		// create each publisher channel: channel that a pub-r reads.
		p.ch = make(chan *Item, pl.chanSize)
		pub := p
		go func() {
			defer pl.wg.Done()
			pub.Publish(pub.ch)
		}()
	}

	for _, _s := range pl.sources {
		s, info := _s, _s.Info()
		sink := info.newSink(pl.prodc)
		// add rotator element for each source
		pl.rot.Elems = append(pl.rot.Elems, rotatorElem{
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
		pl.rot.run(pl.quit)
	})
	GoWG(&pl.wg, pl.run)
	return nil
}

func (pl *Pipeline) run() {

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

		if !pl.dedup.Keep(it.key) {
			continue
		}

		for pname := range pubs {
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

//Stop - stops pipeline
//todo check & refac.
func (pl *Pipeline) Stop() {
	pl.quit <- struct{}{}
	close(pl.prodc)
	for _, p := range pl.pubs {
		close(p.ch)
	}
}

// todo refac/remove
func (pl *Pipeline) Wait() {
	pl.wg.Wait()
}
