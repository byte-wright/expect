package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestNotToBe(t *testing.T) {
	expect.Value(t, "number", 7).NotToBe(8)
}

func TestFailNotToBe(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "number", 7).NotToBe(7)
	})
	l.ExpectMessage(0).ToBe("expected number to NOT be 7 but it is")
}

func TestFailNotToBeSlice(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "numbers", []int{3, 2, 1}).NotToBe([]int{3, 2, 1})
	})
	l.ExpectMessage(0).ToBe(`expected numbers to NOT be
- 3
- 2
- 1
but it is`)
}

func TestNotToBeNil(t *testing.T) {
	expect.Value(t, "pointer", &struct{}{}).NotToBe(nil)
	expect.Value(t, "list", []int{}).NotToBe(nil)
}

func TestFailNotToBeNil(t *testing.T) {
	type user struct{}

	var (
		u    *user
		list []int
		dict map[string]int
	)

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "user", u).NotToBe(nil)
		expect.Value(t, "list", list).NotToBe(nil)
		expect.Value(t, "map", dict).NotToBe(nil)
	})
	l.ExpectMessages().ToCount(3)
	l.ExpectMessage(0).ToBe("expected user to NOT be nil but it is")
	l.ExpectMessage(1).ToBe("expected list to NOT be nil but it is")
	l.ExpectMessage(2).ToBe("expected map to NOT be nil but it is")
}
