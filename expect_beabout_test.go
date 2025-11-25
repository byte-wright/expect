package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestToBeAbout(t *testing.T) {
	expect.Value(t, "liters", 1.92).ToBeAbout(2, 0.1)
	expect.Value(t, "liters", float32(1.92)).ToBeAbout(2, 0.1)
	expect.Value(t, "liters", 98).ToBeAbout(100, 5)
}

func TestFailToBeAbout(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "liters", 1.92).ToBeAbout(2, 0.01)
		expect.Value(t, "liters", 98).ToBeAbout(100, 1)
	})
	l.ExpectMessage(0).ToBe("expected liters to be 2±0.01 but it is 1.92")
	l.ExpectMessage(1).ToBe("expected liters to be 100±1 but it is 98")
}
