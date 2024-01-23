package batchr

import (
	"reflect"
	"testing"
)

/*******************************
 * Test Helper Functions Below *
 *******************************/

const tFail = "test failed"

func failTest(t *testing.T, msgAndArgs0 ...any) {
	t.Helper()

	if len(msgAndArgs0) == 0 {
		t.Errorf(tFail)
	}
	var msg string
	var msgOK bool
	msg, msgOK = msgAndArgs0[0].(string)
	if !msgOK {
		t.Errorf(tFail)
	}
	args := msgAndArgs0[1:]
	t.Errorf(msg, args...)
}

func assrtEqual(t *testing.T, oe, oa any, msgAndArgs0 ...any) {
	t.Helper()

	if reflect.DeepEqual(oe, oa) {
		return
	}

	if len(msgAndArgs0) == 0 {
		t.Errorf("expected values to be equal, but\n\texpected value was %v\n\tactual value was %v", oe, oa)
		return
	}
	failTest(t, msgAndArgs0...)
}

func assrtEqualAny(t *testing.T, oe0 []any, oa any, msgAndArgs0 ...any) {
	t.Helper()
	if len(oe0) == 0 {
		t.Errorf("impossible - empty assertion for EqualAny")
		return
	}
	for _, oe := range oe0 {
		if oe == oa {
			return
		}
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected value to equal one of %v, but\n\tactual value was %v", oe0, oa)
		return
	}
	failTest(t, msgAndArgs0...)
}

func assrtFalse(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == false {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected bool false, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0...)
}

func assrtTrue(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == true {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected bool true, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0...)
}

func assrtNil(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o == nil {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected nil, but actual value was %v", o)
		return
	}
	failTest(t, msgAndArgs0...)
}

func assrtNotNil(t *testing.T, o any, msgAndArgs0 ...any) {
	t.Helper()

	if o != nil {
		return
	}
	if len(msgAndArgs0) == 0 {
		t.Errorf("expected non-nil, but actual value was nil")
		return
	}
	failTest(t, msgAndArgs0...)
}
