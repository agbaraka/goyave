package ratelimiter

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v3"
)

func TestNewLimiter(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	assert.Equal(t, 5, l.config.RequestQuota)
	assert.Equal(t, 0, l.counter)
	assert.NotEqual(t, int64(0), l.resetsAt.UnixNano())
}

func TestLimiterHasExceededRequestQuota(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.counter = 4
	assert.False(t, l.hasExceededRequestQuota())

	l.counter = 5
	assert.True(t, l.hasExceededRequestQuota())

	l.counter = 6
	assert.True(t, l.hasExceededRequestQuota())
}

func TestLimiterGetRemainingRequestQuota(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.counter = 3
	assert.Equal(t, 2, l.getRemainingRequestQuota())
}

func TestLimiterStore(t *testing.T) {
	store := newLimiterStore()
	assert.NotNil(t, store.store)

	config := Config{
		QuotaDuration: 500 * time.Millisecond,
	}
	l := newLimiter(config)
	store.set("key", l)
	limiter, ok := store.store["key"]
	assert.True(t, ok)
	assert.Same(t, l, limiter)
	assert.Same(t, l, store.get("key", config))

	// Entry should be removed after quota duration expired
	time.Sleep(l.config.QuotaDuration + 10*time.Millisecond)

	newL := store.get("key", config)
	assert.NotNil(t, newL)
	assert.NotSame(t, l, newL)
}

func TestLimiterValidateAndUpdate(t *testing.T) {
	suite := new(goyave.TestSuite)
	l := &limiter{
		config: Config{
			RequestQuota:  5,
			QuotaDuration: time.Second,
		},
		counter:  0,
		resetsAt: time.Now().Add(time.Second),
	}
	valid := l.validateAndUpdate(suite.CreateTestResponse(httptest.NewRecorder()))

	assert.True(t, valid)
	assert.Equal(t, 1, l.counter)

	l.counter = 5
	valid = l.validateAndUpdate(suite.CreateTestResponse(httptest.NewRecorder()))

	assert.False(t, valid)
	assert.Equal(t, 6, l.counter)
}
