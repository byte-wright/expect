package expect_test

import (
	"math"
	"testing"
	"time"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestToBeAboutScalar(t *testing.T) {
	expect.Value(t, "liters", 1.92).ToBeAbout(2.0, expect.FloatDelta(0.1))
	expect.Value(t, "liters", float32(1.92)).ToBeAbout(float32(2), expect.FloatDelta(0.1))
	expect.Value(t, "liters", 98).ToBeAbout(100, expect.IntDelta(5))
}

func TestFailToBeAboutScalar(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "liters", 1.92).ToBeAbout(2.0, expect.FloatDelta(0.01))
		expect.Value(t, "liters", 98).ToBeAbout(100, expect.IntDelta(1))
	})
	l.ExpectMessage(0).ToBe("expected liters to be 2±0.01 but it is 1.92")
	l.ExpectMessage(1).ToBe("expected liters to be 100±1 but it is 98")
}

func TestToBeAboutNaNAndInf(t *testing.T) {
	expect.Value(t, "liters", math.NaN()).ToBeAbout(math.NaN(), expect.FloatDelta(0.1))
	expect.Value(t, "liters", math.Inf(1)).ToBeAbout(math.Inf(1), expect.FloatDelta(0.1))
}

func TestFailToBeAboutNaN(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "liters", math.NaN()).ToBeAbout(2.0, expect.FloatDelta(0.1))
		expect.Value(t, "liters", 2.0).ToBeAbout(math.NaN())
	})
	l.ExpectMessage(0).ToBe("expected liters to be 2±0.1 but it is NaN")
	l.ExpectMessage(1).ToBe("expected liters to be NaN±0 but it is 2")
}

type measurement struct {
	Name     string
	Ratio    float64
	Count    int
	Created  time.Time
	Took     time.Duration
	Readings []float64
}

func TestToBeAboutStruct(t *testing.T) {
	base := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	want := measurement{
		Name:     "tank",
		Ratio:    0.5,
		Count:    100,
		Created:  base,
		Took:     time.Second,
		Readings: []float64{1.0, 2.0, 3.0},
	}
	got := measurement{
		Name:     "tank",
		Ratio:    0.503,
		Count:    102,
		Created:  base.Add(200 * time.Millisecond),
		Took:     time.Second + 5*time.Millisecond,
		Readings: []float64{1.001, 1.999, 3.0},
	}

	expect.Value(t, "measurement", got).ToBeAbout(want,
		expect.FloatDelta(0.01),
		expect.IntDelta(5),
		expect.TimeDelta(time.Second),
		expect.DurationDelta(10*time.Millisecond),
	)
}

func TestFailToBeAboutNestedFloat(t *testing.T) {
	base := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	want := measurement{Name: "tank", Ratio: 0.5, Created: base, Readings: []float64{1.0, 2.0}}
	got := measurement{Name: "tank", Ratio: 0.5, Created: base, Readings: []float64{1.0, 2.5}}

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "measurement", got).ToBeAbout(want, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected measurement.Readings[1] to be 2±0.1 but it is 2.5")
}

func TestFailToBeAboutTime(t *testing.T) {
	want := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	got := want.Add(5 * time.Second)

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "created", got).ToBeAbout(want, expect.TimeDelta(time.Second))
	})
	l.ExpectMessage(0).ToBe("expected created to be 2026-06-30T12:00:00Z±1s but it is 2026-06-30T12:00:05Z")
}

func TestFailToBeAboutZeroTime(t *testing.T) {
	want := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "created", time.Time{}).ToBeAbout(want, expect.TimeDelta(time.Second))
	})
	l.ExpectMessage(0).ToBe("expected created to be 2026-06-30T12:00:00Z±1s but it is 0001-01-01T00:00:00Z")
}

func TestFailToBeAboutDuration(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "took", 2*time.Second).ToBeAbout(time.Second, expect.DurationDelta(10*time.Millisecond))
	})
	l.ExpectMessage(0).ToBe("expected took to be 1s±10ms but it is 2s")
}

func TestFailToBeAboutInt(t *testing.T) {
	want := measurement{Name: "tank", Count: 100}
	got := measurement{Name: "tank", Count: 110}

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "measurement", got).ToBeAbout(want, expect.IntDelta(5))
	})
	l.ExpectMessage(0).ToBe("expected measurement.Count to be 100±5 but it is 110")
}

func TestFailToBeAboutIntOverflow(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "limit", int64(math.MaxInt64)).ToBeAbout(int64(-1))
		expect.Value(t, "limit", int64(math.MinInt64)).ToBeAbout(int64(math.MaxInt64), expect.IntDelta(5))
		expect.Value(t, "took", time.Duration(math.MaxInt64)).ToBeAbout(time.Duration(-1))
	})
	l.ExpectMessage(0).ToBe("expected limit to be -1±0 but it is 9223372036854775807")
	l.ExpectMessage(1).ToBe("expected limit to be 9223372036854775807±5 but it is -9223372036854775808")
	l.ExpectMessage(2).ToBe("expected took to be -1ns±0s but it is 2562047h47m16.854775807s")
}

func TestFailToBeAboutNegativeDelta(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "count", uint(5)).ToBeAbout(uint(100), expect.IntDelta(-1))
	})
	l.ExpectMessage(0).ToBe("expected count to be 100±-1 but it is 5")
}

func TestFailToBeAboutString(t *testing.T) {
	want := measurement{Name: "tank", Ratio: 0.5}
	got := measurement{Name: "barrel", Ratio: 0.5}

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "measurement", got).ToBeAbout(want, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected measurement.Name to be 'tank' but it is 'barrel'")
}

func TestFailToBeAboutSliceLength(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "readings", []float64{1, 2}).ToBeAbout([]float64{1, 2, 3}, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected readings to have 3 elements but it has 2 elements")
}

func TestFailToBeAboutType(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "liters", 1.8).ToBeAbout(float32(2.0), expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected liters to be of type float32 but it is of type float64")
}
