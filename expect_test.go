package expect_test

import (
	"errors"
	"testing"

	"github.com/byte-wright/expect"
	"github.com/byte-wright/expect/internal/test"
)

func TestExample(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "the guy", "Peter").ToBe("Steven")
	})
	l.ExpectMessage(0).ToBe("expected the guy to be 'Steven' but it is 'Peter'")
}

func TestErrorToHaveMessage(t *testing.T) {
	expect.Error(t, errors.New("I am the error message")).Message().ToBe("I am the error message")
	expect.Value(t, "error", errors.New("I am the error message")).Message().ToBe("I am the error message")
}

func TestNilErrorMessage(t *testing.T) {
	expect.Error(t, nil).Message().ToBe("")
}

func TestIntToNotAllowMessageMethod(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		expect.Value(t, "int", 0).Message().ToBe("0")
	})
	l.ExpectMessage(0).ToBe("Message must only be called on a error value")
}

func TestColoredOutputSpaceChars(t *testing.T) {
	l := test.New(t, func(t expect.Test) {
		o := expect.Default.Output
		expect.Default.Output = expect.ColoredDiffOutput
		expect.Value(t, "spaces", " 	\n----------------------").ToBe("----------------------")
		expect.Default.Output = o
	})
	l.ExpectMessage(0).ToBe("[31m․↦↵\n[0m----------------------")
}
