package cmd

import "errors"

// ErrLintFailure is returned by check when lint issues are found.
// The caller should exit with code 1.
var ErrLintFailure = errors.New("lint failure")
