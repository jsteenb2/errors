# Package errors

`errors` makes go error handling more powerful and expressive
while working alongside the std lib errors. This pkg is largely
inspired by [upspin project's error handling](https://github.com/upspin/upspin/blame/master/errors/errors.go)
and github.com/pkg/errors. It also adds a few lessons learned
from creating a module like this a number of times across numerous
projects.

## Familiar favorites at your disposal

This pkg is a drop in replace for `github.com/pkg/errors`, with a nearly
identical interface. Similarly, the std lib `errors` module's functionality
have been replicated in this pkg so that you only ever have to work from
a single errors module. Here are some examples of what you can do:

```go
package foo

import (
	"github.com/jsteenb2/errors"
)

func Simple() error {
	return errors.New("simple error")
}

func Enriched() error {
	return errors.New("enriched error", errors.KVs("key_1", "some val", "power_level", 9000))
}

func ErrKindInvalid() error {
	return errors.New("invalid kind error", errors.Kind("invalid"))
}

func Wrapped() error {
	// note if errors.Wrap is passed a nil error, then it returns a nil.
	// matching the behavior of github.com/pkg/errors
	return errors.Wrap(Simple())
}

func Unwrapped() error {
	// no need to import multiple errors pkgs to get the std lib behavior. The
	// small API surface area for the std lib errors are available from this module.
	return errors.Unwrap(Wrapped()) // returns simple error again
}

func WrapFields() error {
	// Add an error Kind and some additional KV metadata. Enrich those errors, and better
	// inform the oncall you that might wake up at 03:00 in the morning :upside_face:
	return errors.Wrap(Enriched(), errors.Kind("some_err_kind"), errors.KVs("dodgers", "stink"))
}

func Joined() error {
	// defaults to printing joined/multi errors as hashicorp's go-multierr does. The std libs,
	// formatter can also be provided.
	return errors.Join(Simple(), Enriched(), ErrKindInvalid())
}

func Disjoined() []error {
	// splits up the Joined errors back to their indivisible parts []error{Simple, Enriched, ErrKindInvalid}
	return errors.Disjoin(Joined())
}
```

This is a quick example of what's available. The std lib `errors`, `github.com/pkg/errors`,
hashicorp's `go-multierr`, and the `upspin` projects error handling all bring incredible
examples of error handling. However, they all have their limitations.

The std lib errors are intensely simple. Great for a hot path, but not great for creating
structured/enriched errors that are useful when creating services and beyond.

The `github.com/pkg/errors` laid the ground work for what is the std lib `errors` today, but
also provided access to a callstack for the errors. This module takes a similar approach to
`github.com/pkg/errors`'s callstack capture, except that it is not capturing the entire stack
all the time. We'll touch on this more soon.

Now with the `go-multierr` module, we have excellent ways to combine errors into a single return
type that satisfies the `error` interface. However, similar to the std lib, that's about all
it does. You can use Is/As with it, which is great, but it does not provide any means to add
additional context or behavior.

The best in show (imo, YMMV) for `error` modules is the `upspin` project's error handling. The
obvious downside to it, is its specific to `upspin`. For many applications creating this whole
error handling setup wholesale, can be daunting as the `upspin` project did an amazing job of
writing their `error` pkg to suit their needs.

## Error behavior untangles the error handling hairball

Instead of focusing on a multitude of specific error types or worse,
a gigantic list of sentinel errors, one for each individual "thing",
you can category your errors with `errors.Kind`. The following is an
error that exhibits a `not_found` behavior.

```go
// domain.go

package foo

import (
	"github.com/jsteenb2/errors"
)

const (
	ErrKindInvalid  = errors.Kind("invalid")
	ErrKindNotFound = errors.Kind("not_found")
	// ... additional
)

func FooDo() {
	err := complexDoer()
	if errors.Is(ErrKindNotFound, err) {
		// handle not found error
	}
}

func complexDoer() error {
	// ... trim
	return errors.New("some not found error", ErrKindNotFound)
}
```

This works across any number of functional boundaries. Instead of
creating a new type that just holds info like a NotFound method or
a field for behavior, we can utilize the `errors.Kind` to decorate
our error handling. Imagine a situation you've probably been in before,
a service that has N entities, and error handling that sprawls with
the use of one off error types or worse, a bazillion sentinel errors.
This makes it difficult to abstract across. By categorizing errors,
you can create strong abstractions. Take the http layer, where we often
want to correlate an error to a specific HTTP Status code. With the
kinds above we can do something like:

```go
package foo

import (
	"net/http"
	
	"github.com/jsteenb2/errors"
)

func errHTTPStatus(err error) int {
	switch {
	case errors.Is(ErrKindInvalid, err):
		return http.StatusBadRequest
	case errors.Is(ErrKindNotFound, err):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
```

Pretty neat yah?

## Adding metadata/fields to contextualize the error

One of the strongest cases I can make for this module is the use of the `errors.Fields`
function we provide. Each error created, wrapped or joined, can have additional metadata
added to the error to contextualize the error. Instead of wrapping the error using
`fmt.Errorf("new context str: %w", err)`, you can use intelligent error handling, and
leave the message as refined as you like. Its simpler to just see the code in action:

```go
package foo

import (
	"github.com/jsteenb2/errors"
)

func Up(timeline string, powerLvl, teamSize int) error {
	return errors.New("the up failed", errors.Kind("went_ape"), errors.KVs(
		"timeline", timeline,
		"power_level", powerLvl,
		"team_size", teamSize,
		"rando_other_field", "dodgers stink",
	))
}

func DownUp(timeline string, powerLvl, teamSize int) error {
	// ... trim 
	err := Up(timeline, powerLvl, teamSize)
	return errors.Wrap(err)
}
```

Here we are returning an error from the `Up` function, that has some context
added (via `errors.KVs` and `errors.Kind`). Additionally, we get a stack trace
added as well. Now lets see what these fields actually look like:

```go
package foo

import (
	"fmt"
	
	"github.com/jsteenb2/errors"
)

func do() {
	err := DownUp("dbz", 9009, 4)
	if err != nil {
		fmt.Printf("%#v\n", errors.Fields(err))
		/*
		    Outputs: []any{
		        "timeline", "dbz",
		        "power_level", 9009,
		        "team_size", 4,
		        "rando_other_field", "dodgers stink",
		        "err_kind", "went_ape",
		        "stack_trace", []string{
		            "github.com/jsteenb2/README.go:26[DownUp]",  // the wrapping point
		            "github.com/jsteenb2/README.go:15[Up]", // the new error call
		        },
		    }
		
		   Note: the filename in hte stack trace is made up for this read me doc. In
		         reality, it'll show the file of the call to errors.{New|Wrap|Join}.
		*/
	}
}
```

The above output, is golden for informing logging infrastructure. Becomes very
simple to create as much context as possible to debug an error. It becomes
very easy to follow the advice of John Carmack, by adding assertions, or
good error handling in go's case, without having to drop a blender on the actual
error message. When that error message remains clean, it empowers your observability
stack. Reducing the cardinality and being able to see across the different facets
your fields provide can create opportunities to explore relationships between failures.
Additionally, there's a fair chance that a bunch of `DEBUG` logs can be removed. Your
SRE/infra teams will thank for it :-).

## Limitations

Worth noting here, this pkg has some limitations, like in
hot loops. In that case, you may want to use std lib errors or
similar for the hot loop, then return the result of those with
this module's error handling with a simple:

```go
errors.Wrap(hotPathErr)
```