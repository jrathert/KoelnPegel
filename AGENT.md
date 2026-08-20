# Agent Guide

## Project

KoelnPegel is a small Go command-line bot. It reads the current Cologne Rhine
gauge measurement from PEGELONLINE, decides whether a Mastodon status should
be published, and stores measurement and posting history locally.

## Layout

- Root package: executable flow, configuration, decision rules, history, and Mastodon publishing.
- `wsv/`: PEGELONLINE HTTP client and JSON decoding.
- `elwis/`: experimental prognosis scraper; not part of the executable flow.
- `doc/`: architecture and contributor documentation.

## Development

- Use Go 1.19 or newer.
- Run tests with `go test ./...`.
- Build with `go build -o kpg .`.
- Run locally with `KPG_TEST=1`; this prints the generated status instead of posting it.
- The executable expects `kpg.env` in the working directory. Keep credentials out of Git.

## Change Guidance

- Preserve the existing package split and small public surface.
- Keep network-dependent tests in mind; `wsv` tests call PEGELONLINE directly.
- Test threshold, scheduling, history, and status-text behavior when changing the decision flow.
- Do not commit `.kpg*`, `*.env`, or generated binaries.

See [doc/architecture.md](doc/architecture.md) for the current design and runtime flow.