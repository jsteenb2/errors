package errors_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jsteenb2/errors"
)

func Test_Errors(t *testing.T) {
	type wants struct {
		msg    string
		fields []any
	}

	tests := []struct {
		name  string
		input error
		want  wants
	}{
		{
			name:  "simple new error",
			input: errors.New("simple msg"),
			want: wants{
				msg:    "simple msg",
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:24[Test_Errors]"}},
			},
		},
		{
			name:  "with error kind",
			input: errors.New("simple msg", errors.Kind("tester")),
			want: wants{
				msg:    "simple msg",
				fields: []any{"err_kind", "tester", "stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:32[Test_Errors]"}},
			},
		},
		{
			name:  "with kv pair",
			input: errors.New("simple msg", errors.KV{K: "key1", V: "val1"}),
			want: wants{
				msg:    "simple msg",
				fields: []any{"key1", "val1", "stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:40[Test_Errors]"}},
			},
		},
		{
			name:  "with kv pairs",
			input: errors.New("simple msg", errors.KVs("k1", "v1", "k2", []string{"somevalslc"})),
			want: wants{
				msg:    "simple msg",
				fields: []any{"k1", "v1", "k2", []string{"somevalslc"}, "stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:48[Test_Errors]"}},
			},
		},
		{
			name:  "without stack trace",
			input: errors.New("simple msg", errors.NoFrame),
			want: wants{
				msg:    "simple msg",
				fields: []any{},
			},
		},
		{
			name:  "with New and error to wrap",
			input: errors.New("wrap msg", fmt.Errorf("a std lib error")),
			want: wants{
				msg:    "wrap msg: a std lib error",
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:64[Test_Errors]"}},
			},
		},
		{
			name: "with frame skip",
			input: func() error {
				// should match line 75 (function call execution) in stack trace
				return errors.New("simple msg", errors.SkipCaller)
			}(),
			want: wants{
				msg:    "simple msg",
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:75[Test_Errors]"}},
			},
		},
		{
			name:  "with wrap of std lib error",
			input: errors.Wrap(fmt.Errorf("simple error"), "wrap msg"),
			want: wants{
				msg:    "wrap msg: simple error",
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_test.go:83[Test_Errors]"}},
			},
		},
		{
			name: "with wrap of errors error",
			input: errors.Wrap(
				errors.New("simple error"),
				"wrap msg",
			),
			want: wants{
				msg: "wrap msg: simple error",
				fields: []any{"stack_trace", []string{
					"github.com/jsteenb2/errors/errors_test.go:91[Test_Errors]",
					"github.com/jsteenb2/errors/errors_test.go:92[Test_Errors]",
				}},
			},
		},
		{
			name: "with wrap of errors error and mix of options",
			input: errors.Wrap(
				errors.New("simple error", errors.KVs("inner_k1", "inner_v1")),
				"wrap msg",
				errors.KVs("wrapped_k1", "wrapped_v1"),
			),
			want: wants{
				msg: "wrap msg: simple error",
				fields: []any{
					"wrapped_k1", "wrapped_v1",
					"inner_k1", "inner_v1",
					"stack_trace", []string{
						"github.com/jsteenb2/errors/errors_test.go:105[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:106[Test_Errors]",
					},
				},
			},
		},
		{
			name: "with wrap of errors error with outer error kind",
			input: errors.Wrap(
				errors.New("simple error"),
				errors.Kind("wrapper"),
			),
			want: wants{
				msg: "simple error",
				fields: []any{
					"err_kind", "wrapper",
					"stack_trace", []string{
						"github.com/jsteenb2/errors/errors_test.go:124[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:125[Test_Errors]",
					},
				},
			},
		},
		{
			name: "with wrap of errors error with inner error kind",
			input: errors.Wrap(
				errors.New("simple error", errors.Kind("inner")),
			),
			want: wants{
				msg: "simple error",
				fields: []any{
					"err_kind", "inner",
					"stack_trace", []string{
						"github.com/jsteenb2/errors/errors_test.go:141[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:142[Test_Errors]",
					},
				},
			},
		},
		{
			name: "with multiple wraps of errors error with inner error kind",
			input: errors.Wrap(
				errors.Wrap(
					errors.Wrap(
						errors.New("simple error", errors.Kind("inner")),
					),
				),
			),
			want: wants{
				msg: "simple error",
				fields: []any{
					"err_kind", "inner",
					"stack_trace", []string{
						"github.com/jsteenb2/errors/errors_test.go:157[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:158[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:159[Test_Errors]",
						"github.com/jsteenb2/errors/errors_test.go:160[Test_Errors]",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq(t, tt.want.msg, tt.input.Error())
			eqFields(t, tt.want.fields, errors.Fields(tt.input))
		})
	}
}

func eq[T comparable](t *testing.T, want, got T) bool {
	t.Helper()

	matches := want == got
	if !matches {
		t.Errorf("values do not match:\n\t\twant:\t%#v\n\t\tgot:\t%#v", want, got)
	}
	return matches
}

func eqFields(t *testing.T, want, got []any) bool {
	t.Helper()

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

func isErr(t *testing.T, err error) bool {
	t.Helper()

	matches := err != nil
	if !matches {
		t.Fatalf("expected error:\n\t\tgot:\t%v", err)
	}

	return matches
}

func must(t *testing.T, outcome bool) {
	t.Helper()

	if !outcome {
		t.FailNow()
	}
}
