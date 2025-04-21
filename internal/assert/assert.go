package assert

import (
	"reflect"
	"testing"
)

func NotNil(t *testing.T, v any) {
	if v == nil {
		t.Fatal("value should not be nil")
	}
}

func NoErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("error should be nil: %s", err.Error())
	}
}

func Err(t *testing.T, err error) {
	if err == nil {
		t.Fatal("error should not be nil")
	}
}

func Equals(t *testing.T, expected any, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("values are not equal:\nExpected \n%v\nBut got:\n%v\n", expected, actual)
	}
}
