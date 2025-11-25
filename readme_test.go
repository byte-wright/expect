package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestReadmeToBeString(t *testing.T) {
	l := &test.Logger{}
	expect.Value(l, "the house", "big").ToBe("small")
	expect.Value(t, "error", l.Messages[0]).ToBe("expected the house to be 'small' but it is 'big'")
}

func TestReadmeToToBeArray(t *testing.T) {
	l := &test.Logger{}
	expect.Value(l, "array", []int{3, 1}).ToBe([]int{1, 3})
	expect.Value(t, "error", l.Messages[0]).ToBe(`expected array to be
    - 1
    - 3
but it is
    - 3
    - 1`)
}

func TestReadmeToBeFloat64(t *testing.T) {
	l := &test.Logger{}
	expect.Value(l, "liters", 3.4500000000001).ToBe(3.45)
	expect.Value(t, "error", l.Messages[0]).ToBe("expected liters to be 3.45 but it is 3.4500000000001")
}

func TestReadmeToCountString(t *testing.T) {
	l := &test.Logger{}
	expect.Value(l, "token", "F7gTr7y").ToCount(8)
	expect.Value(t, "error", l.Messages[0]).ToBe("expected token to have 8 elements but it has 7 elements")
}
