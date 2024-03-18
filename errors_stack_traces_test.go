package errors_test

import (
	"fmt"
	"testing"

	"github.com/jsteenb2/errors"
)

/*
	This file exists b/c the below tests do not interoperate well with
	active development as any import statement will throw off the stack
	traces. Isolating them here, allows us to make errors_stack_traces_test.go more
	active, without having to worry about imports dorking up a large
	swathe of tests.
*/

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
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:31[Test_Errors]"}},
			},
		},
		{
			name:  "with error kind",
			input: errors.New("simple msg", errors.Kind("tester")),
			want: wants{
				msg:    "simple msg",
				fields: []any{"err_kind", "tester", "stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:39[Test_Errors]"}},
			},
		},
		{
			name:  "with kv pair",
			input: errors.New("simple msg", errors.KV{K: "key1", V: "val1"}),
			want: wants{
				msg:    "simple msg",
				fields: []any{"key1", "val1", "stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:47[Test_Errors]"}},
			},
		},
		{
			name:  "with kv pairs",
			input: errors.New("simple msg", errors.KVs("k1", "v1", "k2", []string{"somevalslc"})),
			want: wants{
				msg:    "simple msg",
				fields: []any{"k1", "v1", "k2", []string{"somevalslc"}, "stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:55[Test_Errors]"}},
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
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:71[Test_Errors]"}},
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
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:82[Test_Errors]"}},
			},
		},
		{
			name:  "with wrap of std lib error",
			input: errors.Wrap(fmt.Errorf("simple error"), "wrap msg"),
			want: wants{
				msg:    "wrap msg: simple error",
				fields: []any{"stack_trace", []string{"github.com/jsteenb2/errors/errors_stack_traces_test.go:90[Test_Errors]"}},
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
					"github.com/jsteenb2/errors/errors_stack_traces_test.go:98[Test_Errors]",
					"github.com/jsteenb2/errors/errors_stack_traces_test.go:99[Test_Errors]",
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
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:112[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:113[Test_Errors]",
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
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:131[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:132[Test_Errors]",
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
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:148[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:149[Test_Errors]",
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
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:164[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:165[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:166[Test_Errors]",
						"github.com/jsteenb2/errors/errors_stack_traces_test.go:167[Test_Errors]",
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
