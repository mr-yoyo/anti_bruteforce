package domain

type Limiter interface {
	IsAllowed(data AuthData) (bool, error)
	DeleteBucket(data BucketData) (bool, error)
}

type Method interface {
	IsLimitExceeded(limit int64) bool
	IsZeroReached() bool
	Increment()
	GetQuitChan() chan bool
}
