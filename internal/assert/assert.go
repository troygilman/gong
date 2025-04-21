package assert

import (
	"reflect"
	"testing"
)

// NotNil asserts that the given value is not nil.
// If the value is nil, the test will fail with a fatal error.
func NotNil(t *testing.T, v any) {
	if v == nil {
		t.Fatal("value should not be nil")
	}
}

// NoErr asserts that the given error is nil.
// If the error is not nil, the test will fail with a fatal error including the error message.
func NoErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("error should be nil: %s", err.Error())
	}
}

// Err asserts that the given error is not nil.
// If the error is nil, the test will fail with a fatal error.
func Err(t *testing.T, err error) {
	if err == nil {
		t.Fatal("error should not be nil")
	}
}

// Equals asserts that the expected and actual values are deeply equal.
// If they are not equal, the test will fail with a fatal error showing the difference.
func Equals(t *testing.T, expected any, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("values are not equal:\nExpected \n%v\nBut got:\n%v\n", expected, actual)
	}
}
