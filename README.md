# Package errors

`errors` makes go error handling more powerful and expressive
while working alongside the std lib errors. This pkg is largely
inspired by [upspin project's error handling](https://github.com/upspin/upspin/blame/master/errors/errors.go)
and github.com/pkg/errors. It also adds a few lessons learned
from creating a module like this a number of times across numerous
projects.

## Error behavior untangles the error handling hairball

Instead of focusing on a multitude of specific error types or worse,
a gigantic list of sentinel errors, one for each individual "thing",
you can category your errors with `errors.Kind`. The following is an
error that exhibits a `not_found` behavior.

```go
// domain.go

package foo

import (
	stderrors "errors"
	
	"github.com/jsteenb2/errors"
)

const (
	ErrKindInvalid  = errors.Kind("invalid")
	ErrKindNotFound = errors.Kind("not_found")
	// ... additional
)

func FooDo() {
	err := complexDoer()
	if stderrors.Is(ErrKindNotFound, err) {
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
	stderrors "errors"
	"net/http"
)

func errHTTPStatus(err error) int {
	switch {
	case stderrors.Is(ErrKindInvalid, err):
		return http.StatusBadRequest
	case stderrors.Is(ErrKindNotFound, err):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
```

Pretty neat yah?

## Limitations

Worth noting here, this pkg has some limitations, like in
hot loops. In that case, you may want to use std lib errors or
similar for the hot loop, then return the result of those with
this module's error handling with a simple:

```go
errors.Wrap(hotPathErr)
```