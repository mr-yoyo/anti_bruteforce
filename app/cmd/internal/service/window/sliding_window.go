package window

import (
	"math"
	"sync"
	"time"

	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
)

const (
	MilliSecondsToUpdate   = 1000
	MilliSecondsSizeWindow = 60000
)

type slidingWindow struct {
	quit     chan bool
	ticker   *time.Ticker
	lock     *sync.RWMutex
	current  window
	previous window
}

type window struct {
	count   int64
	created int64
}

func NewSlidingWindow() domain.Method {
	now := time.Now()
	s := slidingWindow{
		lock:   new(sync.RWMutex),
		quit:   make(chan bool),
		ticker: time.NewTicker(MilliSecondsToUpdate * time.Millisecond),
		current: window{
			count:   0,
			created: now.UnixMilli(),
		},
		previous: window{
			count:   0,
			created: now.Add(-time.Millisecond * MilliSecondsSizeWindow).UnixMilli(),
		},
	}

	go s.process()

	return &s
}

func (s *slidingWindow) IsLimitExceeded(limit int64) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	msPassed := time.Now().UnixMilli() - s.current.created
	prevWeight := 1 - float64(msPassed)/float64(MilliSecondsSizeWindow)

	if prevWeight < 0 {
		prevWeight = 0
	}

	total := int64(math.Round(float64(s.current.count) + prevWeight*float64(s.previous.count)))

	return total > limit
}

func (s *slidingWindow) IsZeroReached() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.current.count == 0 && s.previous.count == 0
}

func (s *slidingWindow) Increment() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.current.count++
}

func (s *slidingWindow) GetQuitChan() chan bool {
	return s.quit
}

func (s *slidingWindow) newWindow() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.previous = s.current
	s.current = window{
		count:   0,
		created: time.Now().UnixMilli(),
	}
}

func (s *slidingWindow) process() {
	for range s.ticker.C {
		if time.Now().UnixMilli() >= s.current.created+MilliSecondsSizeWindow {
			s.newWindow()
		}

		if s.IsZeroReached() {
			s.quit <- true
			close(s.quit)
			return
		}
	}
}
