// Package version exposes the build version of the engine, injected at build
// time via -ldflags "-X dungeons/internal/version.Version=...". It defaults to
// "dev" for `go run` / un-stamped builds.
package version

// Version is the engine build identifier (typically a short git hash, with a
// "-dirty" suffix when built from an uncommitted working tree).
var Version = "dev"
