package errors_test

import (
	"fmt"
	"testing"

	"github.com/jsteenb2/errors"
)

var sentinelErr = fmt.Errorf("sentinel err")

func TestJoin(t *testing.T) {
	t.Run("single error joined error can be unwrapped", func(t *testing.T) {
		err := errors.Join(errors.New("first multi error"))

		gotMsg := err.Error()
		eq(t, "1 error occurred:\n\t* first multi error\n", gotMsg)

		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			t.Fatal("unexpected nil unwrapped error")
		}

		gotMsg = unwrappedErr.Error()
		eq(t, "first multi error", gotMsg)
	})

	t.Run("multiple joined errors can be unwrapped", func(t *testing.T) {
		err := errors.Join(
			errors.New("err 1"),
			errors.New("err 2"),
		)

		wantMsg := `2 errors occurred:
	* err 1
	* err 2
`
		eq(t, wantMsg, err.Error())

		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			t.Fatal("unexpected nil unwrapped error")
		}
		eq(t, "err 1", unwrappedErr.Error())

		unwrappedErr = errors.Unwrap(unwrappedErr)
		if unwrappedErr == nil {
			t.Fatal("unexpected nil unwrapped error")
		}
		eq(t, "err 2", unwrappedErr.Error())
	})

	t.Run("multiple joined errors can be used with Is and As", func(t *testing.T) {
		err := errors.Join(
			errors.New("err 1", errors.Kind("foo")),
			sentinelErr,
		)

		wantMsg := `2 errors occurred:
	* err 1
	* sentinel err
`
		eq(t, wantMsg, err.Error())

		if !errors.Is(err, sentinelErr) {
			t.Errorf("failed to identify sentinel error")
		}
		if !errors.Is(err, errors.Kind("foo")) {
			t.Error("failed to find matching kind error")
		}
	})

	t.Run("multiple joined errors can be used with Fields", func(t *testing.T) {
		err := errors.Join(
			errors.New("err 1", errors.Kind("foo"), errors.KVs("ki1", "vi1")),
			sentinelErr,
			errors.New("err 3", errors.KVs("ki3", "vi3")),
			errors.Join(
				errors.New("err 4"),
			),
			errors.KVs("kj1", "vj1"),
		)
		wantFields := []any{
			// parent Join error
			"kj1", "vj1", "err_kind", "foo", "stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:74[TestJoin.func4]"},
			// first err
			"err_0", []any{"ki1", "vi1", "err_kind", "foo", "stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:75[TestJoin.func4]"}},
			// third err
			"err_2", []any{"ki3", "vi3", "stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:77[TestJoin.func4]"}},
			// fourth err
			"err_3", []any{
				"stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:78[TestJoin.func4]"},
				"err_0", []any{"stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:79[TestJoin.func4]"}},
			},
		}
		eqFields(t, wantFields, errors.Fields(err))

		unwrapped := errors.Unwrap(err)
		wantFields = []any{"ki1", "vi1", "err_kind", "foo", "stack_trace", []string{"github.com/jsteenb2/errors/join_test.go:75[TestJoin.func4]"}}
		eqFields(t, wantFields, errors.Fields(unwrapped))

		sentinelUnwrapped := errors.Unwrap(unwrapped)
		eqFields(t, nil, errors.Fields(sentinelUnwrapped))
	})
}
