//go:build race

package goop

// raceEnabled reports whether the race detector is on. Wall-clock assertions skip under it,
// because instrumentation overhead varies by Go release and runner.
const raceEnabled = true
