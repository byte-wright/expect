package expect_test

import (
	"fmt"
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

func TestToBeAboutInf(t *testing.T) {
	expect.Value(t, "liters", math.Inf(1)).ToBeAbout(math.Inf(1), expect.FloatDelta(0.1))
}

func TestFailToBeAboutNaN(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "liters", math.NaN()).ToBeAbout(2.0, expect.FloatDelta(0.1))
		expect.Value(t, "liters", 2.0).ToBeAbout(math.NaN())
		expect.Value(t, "liters", math.NaN()).ToBeAbout(math.NaN(), expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected liters to be 2±0.1 but it is NaN")
	l.ExpectMessage(1).ToBe("expected liters to be NaN but it is 2")
	l.ExpectMessage(2).ToBe("expected liters to be NaN±0.1 but it is NaN")
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
	l.ExpectMessage(0).ToBe("expected limit to be -1 but it is 9223372036854775807")
	l.ExpectMessage(1).ToBe("expected limit to be 9223372036854775807±5 but it is -9223372036854775808")
	l.ExpectMessage(2).ToBe("expected took to be -1ns but it is 2562047h47m16.854775807s")
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

type node struct {
	Value float64
	Next  *node
}

func ring(values ...float64) *node {
	first := &node{Value: values[0]}
	last := first

	for _, v := range values[1:] {
		last.Next = &node{Value: v}
		last = last.Next
	}

	last.Next = first

	return first
}

func TestToBeAboutCycle(t *testing.T) {
	expect.Value(t, "ring", ring(1.01, 1.99)).ToBeAbout(ring(1, 2), expect.FloatDelta(0.1))
}

func TestFailToBeAboutCycle(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "ring", ring(1, 2.5)).ToBeAbout(ring(1, 2), expect.FloatDelta(0.1))
		expect.Value(t, "ring", ring(1, 2, 3)).ToBeAbout(ring(1, 2), expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected ring.Next.Value to be 2±0.1 but it is 2.5")
	l.ExpectMessage(1).ToBe("expected ring.Next.Next.Value to be 1±0.1 but it is 3")
}

func selfContaining(ratio float64) (map[string]interface{}, []interface{}) {
	dict := map[string]interface{}{"ratio": ratio}
	dict["self"] = dict

	list := []interface{}{ratio, nil}
	list[1] = list

	return dict, list
}

func TestToBeAboutSelfContaining(t *testing.T) {
	wantDict, wantList := selfContaining(0.5)
	gotDict, gotList := selfContaining(0.51)

	expect.Value(t, "dict", gotDict).ToBeAbout(wantDict, expect.FloatDelta(0.1))
	expect.Value(t, "list", gotList).ToBeAbout(wantList, expect.FloatDelta(0.1))
}

func TestFailToBeAboutSelfContaining(t *testing.T) {
	wantDict, wantList := selfContaining(0.5)
	gotDict, gotList := selfContaining(0.7)

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "dict", gotDict).ToBeAbout(wantDict, expect.FloatDelta(0.1))
		expect.Value(t, "list", gotList).ToBeAbout(wantList, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected dict['ratio'] to be 0.5±0.1 but it is 0.7")
	l.ExpectMessage(1).ToBe("expected list[0] to be 0.5±0.1 but it is 0.7")
}

func TestFailToBeAboutSlicesOfOneArray(t *testing.T) {
	type window struct {
		Head, All []float64
	}

	want := []float64{1, 2}
	got := []float64{1, 5}

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "window", window{Head: got[:1], All: got}).ToBeAbout(window{Head: want[:1], All: want})
	})
	l.ExpectMessage(0).ToBe("expected window.All[1] to be 2 but it is 5")
}

func TestToBeAboutNil(t *testing.T) {
	var ratio *float64

	expect.Value(t, "ratio", ratio).ToBeAbout(nil, expect.FloatDelta(0.1))
	expect.Value(t, "ratio", nil).ToBeAbout(ratio)
	expect.Value(t, "readings", []float64(nil)).ToBeAbout(nil)
	expect.Value(t, "ratio", nil).ToBeAbout(nil)
}

func TestFailToBeAboutNil(t *testing.T) {
	ratio := 0.5

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "ratio", &ratio).ToBeAbout(nil, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected ratio to be of type <nil> but it is of type *float64")
}

func TestFailToBeAboutNilVsEmpty(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "readings", []float64(nil)).ToBeAbout([]float64{})
		expect.Value(t, "stock", map[string]int{}).ToBeAbout(map[string]int(nil))
		expect.Value(t, "measurement", measurement{}).ToBeAbout(measurement{Readings: []float64{}})
	})
	l.ExpectMessage(0).ToBe("expected readings to be [] but it is nil")
	l.ExpectMessage(1).ToBe("expected stock to be nil but it is map[]")
	l.ExpectMessage(2).ToBe("expected measurement.Readings to be [] but it is nil")
}

func TestToBeAboutTimeDeltaComparesInstants(t *testing.T) {
	want := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	now := time.Now()

	expect.Value(t, "created", want.In(time.FixedZone("CEST", 2*60*60))).ToBeAbout(want, expect.TimeDelta(0))
	expect.Value(t, "created", now).ToBeAbout(now.Round(0), expect.TimeDelta(0))
}

func TestFailToBeAboutTimeLocation(t *testing.T) {
	want := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "created", want.In(time.FixedZone("CEST", 2*60*60))).ToBeAbout(want)
	})
	l.ExpectMessage(0).ToBe("expected created to be 2026-06-30T12:00:00Z but it is 2026-06-30T14:00:00+02:00" +
		" (same instant, TimeDelta would ignore location and monotonic clock)")
}

type sensor struct {
	Threshold *float64
	ratio     float64
	threshold *float64
	readings  []float64
}

func TestToBeAboutUnexported(t *testing.T) {
	got, want := 1.05, 1.0

	expect.Value(t, "sensor", sensor{Threshold: &got, readings: []float64{got}}).
		ToBeAbout(sensor{Threshold: &want, readings: []float64{want}}, expect.FloatDelta(0.1))
}

func TestFailToBeAboutUnexported(t *testing.T) {
	got, want := 1.05, 1.0

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "sensor", sensor{ratio: got}).ToBeAbout(sensor{ratio: want}, expect.FloatDelta(0.1))
		expect.Value(t, "sensor", sensor{threshold: &got}).ToBeAbout(sensor{threshold: &want}, expect.FloatDelta(0.1))
	})
	l.ExpectMessage(0).ToBe("expected sensor.ratio to be 1 but it is 1.05")
	l.ExpectMessage(1).ToBe("expected sensor.threshold to be 1 but it is 1.05")
}

func TestToBeAboutWithoutOptionsIsToBe(t *testing.T) {
	created := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	nan := []float64{math.NaN()}
	handler := &struct{ OnChange func() }{OnChange: func() {}}

	pairs := []struct{ got, want interface{} }{
		{1.5, 1.5},
		{1.5, 1.6},
		{math.NaN(), math.NaN()},
		{nan, nan},
		{[]float64(nil), []float64{}},
		{map[string]int(nil), map[string]int{}},
		{[]interface{}{nil, 1}, []interface{}{nil, 1}},
		{created, created},
		{created.In(time.FixedZone("CEST", 2*60*60)), created},
		{handler, handler},
		{sensor{ratio: 1}, sensor{ratio: 1}},
		{ring(1, 2), ring(1, 2)},
		{ring(1, 2), ring(1, 3)},
	}

	for i, p := range pairs {
		toBe := test.New(t, func(t expect.Test) { expect.Value(t, "value", p.got).ToBe(p.want) })
		about := test.New(t, func(t expect.Test) { expect.Value(t, "value", p.got).ToBeAbout(p.want) })

		expect.Value(t, fmt.Sprintf("ToBeAbout passing pair %d", i), len(about.Messages) == 0).
			ToBe(len(toBe.Messages) == 0)
	}
}
