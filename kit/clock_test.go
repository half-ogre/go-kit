package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("defaults_to_real_time_when_no_options_provided", func(t *testing.T) {
		beforeTime := time.Now()

		clk := NewClock()
		clockTime := clk.Now()

		afterTime := time.Now()
		assert.True(t, clockTime.After(beforeTime) || clockTime.Equal(beforeTime))
		assert.True(t, clockTime.Before(afterTime) || clockTime.Equal(afterTime))
	})

	t.Run("returns_fake_time_when_with_fake_option_provided", func(t *testing.T) {
		theFixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		fakeTimeFunc := func() time.Time { return theFixedTime }

		clk := NewClock(WithFake(fakeTimeFunc))

		clockTime := clk.Now()
		assert.True(t, clockTime.Equal(theFixedTime))
	})

	t.Run("calls_fake_function_on_each_now_call", func(t *testing.T) {
		callCount := 0
		fakeTimeFunc := func() time.Time {
			callCount++
			return time.Date(2025, 1, 1, 12, callCount, 0, 0, time.UTC)
		}

		clk := NewClock(WithFake(fakeTimeFunc))

		firstTime := clk.Now()
		secondTime := clk.Now()
		expectedFirstTime := time.Date(2025, 1, 1, 12, 1, 0, 0, time.UTC)
		expectedSecondTime := time.Date(2025, 1, 1, 12, 2, 0, 0, time.UTC)
		assert.True(t, firstTime.Equal(expectedFirstTime))
		assert.True(t, secondTime.Equal(expectedSecondTime))
	})
}

func TestWithFake(t *testing.T) {
	t.Run("sets_fake_function_on_clock", func(t *testing.T) {
		theFixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		fakeTimeFunc := func() time.Time { return theFixedTime }
		clk := &clock{}

		option := WithFake(fakeTimeFunc)
		option(clk)

		clockTime := clk.Now()
		assert.True(t, clockTime.Equal(theFixedTime))
	})
}
