package dodod

import (
	"errors"
	"testing"
)

func TestIsErrorType(t *testing.T) {
	t.Helper()
	v := errors.New("custom error")
	if !IsErrorType(v) {
		t.Fatalf("value should a error type")
	}

	v2 := 4
	if IsErrorType(v2) {
		t.Fatalf("value should not be a error type")
	}

}
