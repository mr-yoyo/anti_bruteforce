package window

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_slidingWindow_process(t *testing.T) {
	type fields struct {
		quit     chan bool
		ticker   *time.Ticker
		lock     *sync.RWMutex
		current  window
		previous window
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "new window", fields: fields{
			quit:     make(chan bool),
			ticker:   time.NewTicker(time.Millisecond * 10),
			lock:     new(sync.RWMutex),
			current:  window{3, time.Now().Add(-time.Millisecond * MilliSecondsSizeWindow).UnixMilli()},
			previous: window{2, 0},
		}},
		{name: "zero reach", fields: fields{
			quit:     make(chan bool),
			ticker:   time.NewTicker(time.Millisecond * 10),
			lock:     new(sync.RWMutex),
			current:  window{1, time.Now().Add(-time.Millisecond * MilliSecondsSizeWindow).UnixMilli()},
			previous: window{0, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &slidingWindow{
				quit:     tt.fields.quit,
				ticker:   tt.fields.ticker,
				lock:     tt.fields.lock,
				current:  tt.fields.current,
				previous: tt.fields.previous,
			}

			wg := new(sync.WaitGroup)

			go func() {
				if tt.name == "new window" {
					go s.process()

					wg.Add(1)
					go func() {
						defer wg.Done()
						<-s.ticker.C
					}()

					wg.Wait()

					require.True(t, s.previous.count == tt.fields.current.count)
					require.True(t, s.current.count == 0)
				}

				if tt.name == "zero reach" {
					go s.process()
					var boolQuit bool

					wg.Add(1)
					go func() {
						defer wg.Done()
						boolQuit = <-s.quit
					}()

					wg.Wait()

					require.True(t, s.current.count == 0)
					require.True(t, s.previous.count == 0)
					require.True(t, boolQuit)
				}
			}()
		})
	}
}

func Test_slidingWindow_Increment(t *testing.T) {
	s := &slidingWindow{
		quit:     make(chan bool),
		ticker:   time.NewTicker(time.Millisecond * 10),
		lock:     new(sync.RWMutex),
		current:  window{9, 0},
		previous: window{9, 0},
	}

	s.Increment()

	require.True(t, s.current.count == 10)
	require.True(t, s.previous.count == 9)
}

func Test_slidingWindow_IsLimitExceeded(t *testing.T) {
	limit := int64(10)

	tests := []struct {
		name               string
		current            int64
		previous           int64
		millisecondsPassed time.Duration
		isLimitExceeded    bool
	}{
		{
			name:               "0 ms passed, previous window has weight 1",
			current:            0,
			previous:           10,
			millisecondsPassed: 0 * time.Millisecond,
			isLimitExceeded:    false,
		},
		{
			name:               "0 ms passed, previous window has weight 1, current window counter is 1",
			current:            1,
			previous:           10,
			millisecondsPassed: 0 * time.Millisecond,
			isLimitExceeded:    true,
		},
		{
			name:               "6000 ms passed, previous window has weight 0.9, current window counter is 0",
			current:            0,
			previous:           11,
			millisecondsPassed: 6000 * time.Millisecond,
			isLimitExceeded:    false,
		},
		{
			name:               "6000 ms passed, previous window has weight 0.9, current window counter is 1",
			current:            1,
			previous:           11,
			millisecondsPassed: 6000 * time.Millisecond,
			isLimitExceeded:    true,
		},
		{
			name:               "54000 ms passed, previous window has weight 0.1, current window counter is 0",
			current:            0,
			previous:           100,
			millisecondsPassed: 54000 * time.Millisecond,
			isLimitExceeded:    false,
		},
		{
			name:               "54000 ms passed, previous window has weight 0.1, current window counter is 1",
			current:            1,
			previous:           100,
			millisecondsPassed: 54000 * time.Millisecond,
			isLimitExceeded:    true,
		},
		{
			name:               "59999 ms passed, previous window has weight 0, current window counter is 9",
			current:            9,
			previous:           100,
			millisecondsPassed: 59999 * time.Millisecond,
			isLimitExceeded:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &slidingWindow{
				quit:     make(chan bool),
				ticker:   time.NewTicker(time.Second * 1),
				lock:     new(sync.RWMutex),
				current:  window{tt.current, time.Now().Add(-tt.millisecondsPassed).UnixMilli()},
				previous: window{tt.previous, 0},
			}

			require.Equal(t, s.IsLimitExceeded(limit), tt.isLimitExceeded)
		})
	}
}
