package expect_test

import (
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestToHavePrefix(t *testing.T) {
	expect.Value(t, "statement", "we are all crazy").ToHavePrefix("we are")
}

func TestFailToHavePrefix(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "statement", "we are all crazy").ToHavePrefix("i am")
	})
	l.ExpectMessage(0).ToBe("expected statement to have prefix 'i am' but it is 'we are all crazy'")
}

func TestErrorToHavePrefixOnInt(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "number", 7).ToHavePrefix("i am")
	})
	l.ExpectMessage(0).ToBe("ToHavePrefix must only be called on a string value")
}

func TestToHaveSuffix(t *testing.T) {
	expect.Value(t, "statement", "we are all crazy").ToHaveSuffix("all crazy")
}

func TestFailToHaveSufix(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "statement", "we are all crazy").ToHaveSuffix("all nuts")
	})
	l.ExpectMessage(0).ToBe("expected statement to have suffix 'all nuts' but it is 'we are all crazy'")
}

func TestErrorToHaveSufixOnInt(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "number", 7).ToHaveSuffix("i am")
	})
	l.ExpectMessage(0).ToBe("ToHaveSuffix must only be called on a string value")
}
