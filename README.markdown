`ps` is a Go package to list system processes.

Import as `zgo.at/ps`. Godoc: https://pkg.go.dev/zgo.at/ps#pkg-overview

Supported systems: Linux, Windows, macOS, and FreeBSD.

---

This is a fork of https://github.com/mitchellh/go-ps with the following changes:

- `Processes()` → `List()`; `FindProcess()` → `Find()`; `Process.PPid() →
  `Process.ParentPid()`.

- Add `String()` method, and `List()` returns a `Processes` with `String()`.

- Return an error from `Find()` if a process doesn't exist instead of `nil, nil`.

- Set up test runners for all supported platforms, and remove Solaris/illumos
  support as I can't get that to work as I'm not very familiar with these
  systems. Feel free to fix that and we can add it back, but I'd rather remove
  support for now instead of shipping an untested platform that may or may not
  work.
