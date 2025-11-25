package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestExpectFirstInSlice(t *testing.T) {
	expect.Value(t, "int slice", []int{1, 2, 3}).First().ToBe(1)
}

func TestExpectLastInSlice(t *testing.T) {
	expect.Value(t, "int slice", []int{1, 2, 3}).Last().ToBe(3)
}

func TestFailLastInEmptySlice(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "int slice", []int{}).Last()
	})
	l.ExpectMessage(0).ToBe("int slice is empty, can not take last element")
}

func TestExpectFirstInString(t *testing.T) {
	expect.Value(t, "string", "Alabama").First().ToBe("A")
	expect.Value(t, "string", "äh").First().ToBe("ä")
	expect.Value(t, "string", "日本のジャガイモ").First().ToBe("日")
}

func TestExpectLastInString(t *testing.T) {
	expect.Value(t, "string", "日本のジャガイモ").Last().ToBe("モ")
}

func TestGetLastInArray(t *testing.T) {
	expect.Value(t, "int array", [2]int{1, 2}).Last().ToBe(2)
}

func TestFailLastInMap(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "map", map[string]string{"a": "1", "b": "c"}).Last()
	})
	l.ExpectMessage(0).ToBe("map is not an indexable datatype")
}
