package projectpath

import (
	"path/filepath"
	"runtime"
)

// For usage of this path variable
// See https://stackoverflow.com/a/58294680/628748
var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../..")
)
