// package errors makes error handling more powerful and expressive
// while working alongside the std lib errors. This pkg is largely
// inspired by the [upspin project's error handling](https://github.com/upspin/upspin/blame/master/errors/errors.go)
// and github.com/pkg/errors. It also adds a few lessons learned
// from creating a module like this a number of times across numerous
// projects. Worth noting here, this pkg has some limitations, like in
// hot loops. In that case, you may want to use std lib errors or
// similar for the hot loop, then return the result of those with
// this module's error handling with a simple:
// 	errors.Wrap(hotPathErr)

package errors
