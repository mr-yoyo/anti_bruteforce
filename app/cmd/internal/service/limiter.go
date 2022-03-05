package service

import (
	"sync"

	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/config"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/infrastructure/db"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/service/window"
)

const (
	IpPrefix       = "ip:"
	LoginPrefix    = "lgn:"
	PasswordPrefix = "pwd:"
)

type limiter struct {
	m             sync.Map
	whiteListRepo domain.IPListRepository
	blackListRepo domain.IPListRepository
	limits        struct {
		ip       int64
		login    int64
		password int64
	}
}

func NewLimiter() domain.Limiter {
	cfg := config.Get()

	return &limiter{
		whiteListRepo: db.NewWhitelist(),
		blackListRepo: db.NewBlacklist(),
		limits: struct {
			ip       int64
			login    int64
			password int64
		}{
			ip:       cfg.GetInt64("policy.same_ip"),
			login:    cfg.GetInt64("policy.same_login"),
			password: cfg.GetInt64("policy.same_password"),
		},
	}
}

func (l *limiter) IsAllowed(data domain.AuthData) (bool, error) {
	// check if is in black list (it has higher priority)
	exists, err := l.blackListRepo.Exists(data.IP)
	if err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	// check if is in white list
	exists, err = l.whiteListRepo.Exists(data.IP)
	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}

	ch := make(chan bool, 3)
	defer close(ch)

	go l.isValueAllowed(IpPrefix+data.IP.String(), l.limits.ip, ch)
	go l.isValueAllowed(LoginPrefix+data.Login, l.limits.login, ch)
	go l.isValueAllowed(PasswordPrefix+data.Password, l.limits.password, ch)

	a1, a2, a3 := <-ch, <-ch, <-ch

	return a1 && a2 && a3, nil
}

func (l *limiter) isValueAllowed(key string, limit int64, ch chan<- bool) {
	w := window.NewSlidingWindow()

	if v, ok := l.m.LoadOrStore(key, w); ok {
		w = v.(domain.Method)
	} else {
		go func() {
			for range w.GetQuitChan() {
				l.m.Delete(key)
				return
			}
		}()
	}

	w.Increment()

	ch <- !w.IsLimitExceeded(limit)
}

func (l *limiter) DeleteBucket(data domain.BucketData) (bool, error) {
	_, loginExists := l.m.LoadAndDelete(LoginPrefix + data.Login)
	_, ipExists := l.m.LoadAndDelete(IpPrefix + data.IP.String())

	return loginExists || ipExists, nil
}
