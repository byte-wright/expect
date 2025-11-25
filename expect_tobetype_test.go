package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestToBeTypeString(t *testing.T) {
	expect.Value(t, "foo", "xxx").ToBeType("")
}

func TestFailToBeTypeString(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "foo", 7).ToBeType("7")
	})
	l.ExpectMessages().ToCount(1)
	l.ExpectMessage(0).ToBe("expected foo to be of type 'string' but it is of type 'int'")
}

func TestToBeTypeStruct(t *testing.T) {
	type tsType struct{ A int }
	expect.Value(t, "foo", &tsType{A: 3}).ToBeType(&tsType{})
}

func TestFailToBeTypeStruct(t *testing.T) {
	type tsTypeA struct {
		A int
	}
	type tsTypeB struct {
		A int
	}

	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "foo", &tsTypeA{A: 3}).ToBeType(&tsTypeB{})
	})
	l.ExpectMessages().ToCount(1)
	l.ExpectMessage(0).ToBe("expected foo to be of type '*expect_test.tsTypeB' but it is of type '*expect_test.tsTypeA'")
}
