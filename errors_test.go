package errors_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jsteenb2/errors"
)

func TestWrap(t *testing.T) {
	t.Run("simple wrapped error is returned when calling std lib errors.Unwrap", func(t *testing.T) {
		baseErr := errors.New("first error")

		if unwrapped := errors.Unwrap(baseErr); unwrapped != nil {
			t.Fatalf("recieved unexpected unwrapped error:\n\t\tgot:\t%v", unwrapped)
		}

		wrappedErr := errors.Wrap(baseErr)
		if unwrapped := errors.Unwrap(wrappedErr); unwrapped == nil {
			t.Fatalf("recieved unexpected nil unwrapped error")
		}
	})

	t.Run("unwrapping nil error should return nil", func(t *testing.T) {
		err := errors.Wrap(nil)
		if err != nil {
			t.Fatalf("recieved unexpected wrapped error:\n\t\tgot:\t%v", err)
		}
	})
}

func TestV(t *testing.T) {
	type foo struct {
		i int
	}

	t.Run("key val pairs are should be accessible", func(t *testing.T) {
		err := errors.New("simple msg", errors.KVs("bool", true, "str", "string", "float", 3.14, "int", 1, "foo", foo{i: 3}))

		eqV(t, err, "bool", true)
		eqV(t, err, "str", "string")
		eqV(t, err, "float", 3.14)
		eqV(t, err, "int", 1)
		eqV(t, err, "foo", foo{i: 3})

		if v := errors.V(err, "non existent"); v != nil {
			t.Errorf("unexpected value returned:\n\t\tgot:\t%#v", v)
		}
	})

	t.Run("when parent error kv pair collides with wrapped error will take parent kv pair", func(t *testing.T) {
		err := errors.New("simple msg", errors.KVs("str", "initial"))
		err = errors.Wrap(err, errors.KVs("str", "wrapped"))

		eqV(t, err, "str", "wrapped")
	})
}

func eq[T comparable](t *testing.T, want, got T) bool {
	t.Helper()

	matches := want == got
	if !matches {
		t.Errorf("values do not match:\n\t\twant:\t%#v\n\t\tgot:\t%#v", want, got)
	}
	return matches
}

func eqV[T comparable](t *testing.T, err error, key string, want T) bool {
	t.Helper()

	got, ok := errors.V(err, key).(T)
	must(t, eq(t, true, ok))
	return eq(t, want, got)
}

func eqFields(t *testing.T, want, got []any) bool {
	t.Helper()

	defer func() {
		if t.Failed() {
			b, _ := json.MarshalIndent(got, "", "  ")
			t.Logf("got: %s", string(b))
		}
	}()

	if matches := eqLen(t, len(want), got); !matches {
		return matches
	}

	matches := true

	// compare pairs and verify keys are strings
	for i := 0; i < len(want); i += 2 {
		wantKey := isT[string](t, want[i])
		gotKey := isT[string](t, got[i])

		vIdx := i + 1
		wantVal, gotVal := want[vIdx], got[vIdx]

		keysMatch := wantKey == gotKey
		valsMatch := reflect.DeepEqual(wantVal, gotVal)

		if !keysMatch || !valsMatch {
			matches = false
			t.Errorf("unexpected fields pair:\n\t\twant:\t%s: %+v\n\t\tgot:\t%s: %+v", wantKey, wantVal, gotKey, gotVal)
		}
	}
	return matches
}

func eqLen[T any](t *testing.T, wantLen int, got []T) bool {
	t.Helper()

	matches := wantLen == len(got)
	if !matches {
		t.Fatalf("expected fields to have length %d:\n\t\tgot len: %d\n\t\tgot val: %v", wantLen, len(got), got)
	}
	return matches
}

func isT[T any](t *testing.T, v any) T {
	t.Helper()

	out, ok := v.(T)
	if !ok {
		t.Fatalf("unexpected type found:\n\t\tgot: %T", v)
	}

	return out
}

func must(t *testing.T, outcome bool) {
	t.Helper()

	if !outcome {
		t.FailNow()
	}
}
