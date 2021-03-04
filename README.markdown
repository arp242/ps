`ps` is a Go package to list system processes.

Import as `zgo.at/ps`. Godoc: https://pkg.go.dev/zgo.at/ps

Supported systems: Linux, Windows, macOS, FreeBSD, Solaris.

---

This is a fork of https://github.com/mitchellh/go-ps with the following changes:

- `Processes()` → `List()`.
- `FindProcess()` -> `Find()`.
- Add `String()` method, and `List()` returns a `Processes` with `String()`.
- Return an error from `Find()` if a process doesn't exist, instead of `nil,
  nil`.
