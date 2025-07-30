package kit

import "time"

type ClockInterface interface {
	Now() time.Time
}

type clock struct {
	fake FakeTimeFunc
}

type ClockOption func(*clock)

func NewClock(opts ...ClockOption) ClockInterface {
	c := &clock{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c clock) Now() time.Time {
	if c.fake != nil {
		return c.fake()
	}
	return time.Now()
}

type FakeTimeFunc func() time.Time

func WithFake(timeFunc FakeTimeFunc) ClockOption {
	return func(c *clock) {
		c.fake = timeFunc
	}
}
