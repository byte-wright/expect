package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestToCountArray(t *testing.T) {
	a := [3]string{"a", "b", "c"}
	expect.Value(t, "array", a).ToCount(3)
}

func TestToCountString(t *testing.T) {
	expect.Value(t, "foo", "xxx").ToCount(3)
	expect.Value(t, "foo", "日本のジャガイモ").ToCount(8)
}

func TestToCountMap(t *testing.T) {
	expect.Value(t, "map", map[int]int{1: 2, 2: 3}).ToCount(2)
}

func TestToCountChannel(t *testing.T) {
	ch := make(chan int, 10)
	ch <- 1
	ch <- 2
	expect.Value(t, "channel", ch).ToCount(2)
	<-ch
	expect.Value(t, "channel", ch).ToCount(1)
}

func TestFailToCountString(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "foo", "xxx").ToCount(1)
	})

	l.ExpectMessages().ToCount(1)
	l.ExpectMessage(0).ToBe("expected foo to have 1 elements but it has 3 elements")
}

func TestErrorToCountInt(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "foo", 2).ToCount(2)
	})
	l.ExpectMessages().ToCount(1)
	l.ExpectMessage(0).ToBe("foo is not a datatype with a length (array, slice, map, chan, string)")
}
