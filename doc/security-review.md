# Security Review

**Review date:** 2026-08-20  
**Dependency review updated:** 2026-08-20

## Scope and Method

This review covered the Go dependency graph, tracked repository content and
Git history, credential handling, all outbound network clients, parsing of
untrusted responses, local state files, logging, and the main posting flow.
The following checks were run from the repository root during the original
dependency upgrade:

```text
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go vet ./...
go test ./...
```

`go vet` passed. An earlier `go test ./...` run passed, but a later rerun
received HTTP `403 Forbidden` from the live WSV endpoint in
`TestFetchWsvJSON`; that test is network-dependent and can fail independently
of local code changes. The dependency scan was rerun after the upgrades and
reported **No vulnerabilities found**.

During the follow-up review on the same date, `go vet ./...` passed and an
uncached `go test -count=1 ./...` run failed because the sandbox could not
resolve the live PEGELONLINE host. `govulncheck` was not installed in that
environment, so the earlier successful vulnerability result was not
independently reproduced during the follow-up. These results describe the
state at the review date and do not guarantee that dependencies remain free
of subsequently disclosed vulnerabilities.

## Dependency Review After Upgrade

The repository has been upgraded to Go `1.27.0` and the module's Go version is
now `1.25.0`. The relevant dependency updates are:

- `golang.org/x/net v0.5.0` -> `v0.58.0`
- `github.com/mattn/go-mastodon v0.0.6` -> `v0.0.13`
- `github.com/gorilla/websocket v1.5.0` -> `v1.5.3`
- `github.com/tomnomnom/linkheader` -> `v0.0.0-20250811210735-e5fe3b51442e`

The updated `govulncheck` scan found no known vulnerabilities in the packages
and standard library reachable from this repository. The previous findings
for `golang.org/x/net` and Go's standard library are therefore resolved by the
upgrades, subject to keeping the toolchain and modules current.

The module metadata command could not fully refresh one dependency from
`proxy.golang.org` because the proxy returned `403 Forbidden`; this did not
prevent the completed vulnerability scan from reporting no findings.

## Application Findings

### Credential file permissions

The local `kpg.env` contained the Mastodon configuration and had mode `0664`
at the time of the follow-up review. That made the credentials readable by
other local users and writable by members of the file's group. Its mode was
corrected to `0600` on 2026-08-20. Deployments should create this file with
owner-only access, verify its ownership, and avoid copying it with broader
permissions.

Reference: [README.md](../README.md#runtime-information).

### Unbounded HTTP response reads

`wsv.fetchWsvJSON` and `elwis.FetchPrognosis` use `io.ReadAll` without a size
limit. A compromised or unexpectedly large upstream response could cause high
memory use or process termination. Use a bounded reader such as
`io.LimitReader` and reject responses that exceed the expected maximum.

References: [wsv/wsv.go](../wsv/wsv.go#L62),
[elwis/elwis.go](../elwis/elwis.go#L57).

### Untrusted response content is logged in full

When WSV JSON decoding fails, the complete response is included in the log.
Together with the unbounded response read, a malicious or faulty upstream can
produce very large logs, consume disk space, or place control characters and
misleading content in operational logs. Log only the error and a small,
sanitized preview or response metadata.

Reference: [wsv/wsv.go](../wsv/wsv.go#L35).

### Missing HTTP timeouts and cancellation

The WSV client creates an `http.Client` without a timeout, and the ELWIS code
uses `http.Get`. A stalled network connection can therefore block the process
indefinitely. Mastodon posting also uses `context.Background` without a
deadline. Use explicit clients with bounded request timeouts and request
contexts with deadlines for every outbound request.

References: [wsv/wsv.go](../wsv/wsv.go#L48),
[elwis/elwis.go](../elwis/elwis.go#L51),
[mastopost.go](../mastopost.go#L23).

### Missing HTTP status validation

The clients read and parse response bodies without first requiring a successful
HTTP status. Error pages or proxy responses can be treated as API data or lead
to confusing failures. Check the status code before reading and parsing the
body.

References: [wsv/wsv.go](../wsv/wsv.go#L54),
[elwis/elwis.go](../elwis/elwis.go#L51).

### Incomplete WSV responses can panic the process

`retrieveCurrentData` initializes the water-level and water-temperature
indexes to `-1`, then indexes the fixed-size series array without checking
that both entries were found. Valid JSON with missing, renamed, duplicated,
or reordered series can therefore cause an index-out-of-range panic. The WSV
model also fixes `TimeSeries` at three entries, silently discarding additional
entries. Validate required series, units, timestamps, and plausible numeric
values before constructing a measurement, and return an error rather than
panicking.

References: [core.go](../core.go#L44), [wsv/wsv.go](../wsv/wsv.go#L26).

### Failed retrieval can enter invalid history

`retrieveCurrentData` turns any WSV error into a zero-value measurement. The
main flow continues and can append that value when history is empty. This can
corrupt trend state and conflicts with the documented expectation that
history records successful API calls. Propagate the retrieval error and stop
the posting/history flow unless a validated measurement was obtained.

References: [core.go](../core.go#L38), [kpg.go](../kpg.go#L14),
[kpg.go](../kpg.go#L41).

### Local state file permissions

`.kpg_last` and `.kpg_history` are written with mode `0664`. These files do not
contain access tokens, but another local user who can modify them could alter
posting cadence or trend calculations. Existing files also retain their old
mode when `os.WriteFile` is called. Use `0600`, correct permissions on existing
files, and use atomic replacement with a temporary file in the same directory.

The current reads and writes follow symbolic links. If an attacker can create
files in the working directory, symlinks can redirect reads or writes to other
files accessible to the bot user. Verify that state and configuration paths
are regular files owned by the expected user, reject symlinks, and keep the
working directory non-writable by untrusted users. Atomic renaming improves
crash consistency but does not by itself establish a trusted directory.

References: [core.go](../core.go#L120),
[history.go](../history.go#L61).

### Configuration parser trust boundary

`readEnvironment` accepts any key before `=` and sets it in the process
environment. It has no key allowlist, whitespace handling, duplicate-key
check, value validation, or scanner-error check, and it ignores `os.Setenv`
errors. Restrict accepted keys, reject duplicates and malformed keys, check
scanner and environment-setting errors, and validate that every required
value is present before making network requests.

Reference: [environment.go](../environment.go#L25).

### Configurable credential destination

The `SERVER` value controls where the Mastodon client sends the access token
and client credentials. A typo or unauthorized change to `kpg.env` can send
those secrets to an unintended host. Parse the URL, require HTTPS, reject URL
userinfo and unexpected components, and consider pinning or allowlisting the
intended Mastodon hostname. Protecting `kpg.env` and its containing directory
remains essential.

Reference: [mastopost.go](../mastopost.go#L14).

### Fatal exits in library-style network code

The WSV request-construction path and several ELWIS parsing/error paths call
`log.Fatal`, which terminates the process immediately and bypasses deferred
cleanup. Although the WSV URL is currently constant and ELWIS is not used by
the executable, untrusted ELWIS content can still trigger process termination
when that exported function is called. Return errors to the caller instead.

References: [wsv/wsv.go](../wsv/wsv.go#L49),
[elwis/elwis.go](../elwis/elwis.go#L46).

### Weak validation of persisted state

History JSON has no input-size limit and decoded entries are not bounded to
`HISTORY_LENGTH` until the next save. Timestamps, ordering, numeric values,
and duplicate entries are trusted. A modified or accidentally corrupted file
can increase memory use or manipulate trend decisions. Bound file reads,
validate and normalize entries after decoding, and reject non-finite or
implausible measurements and timestamps.

References: [history.go](../history.go#L28), [history.go](../history.go#L48).

### Posting behavior at the highest threshold

Above `KATA_02`, `checkIfPostNow` posts every invocation without consulting
the last-post interval. This is documented behavior, but a scheduler loop,
replayed measurement, corrupted level, or faulty upstream can create repeated
posts and API load. Validate measurement freshness and consider deduplicating
by measurement timestamp or ID even when the emergency threshold is active.

Reference: [core.go](../core.go#L145).

## Credential and Repository Checks

- `kpg.env` is ignored by Git and was not tracked.
- No credential values were found by the source and Git-history pattern search
  performed during this review. Pattern searches cannot prove that arbitrary
  or unrecognizable secrets were never committed.
- The local `kpg.env` permissions were corrected from `0664` to `0600` during
  the follow-up review; its contents were not printed or added to Git.
- Runtime credentials are read from environment variables and passed to the
  Mastodon client.
- The configured network endpoints use HTTPS.
- No shell execution or dynamic code loading was found.
- The untracked `kpg.org` file is a Go ELF executable and is not covered by the
  `kpg`-only binary ignore rule. Generated binaries should not be committed;
  provenance and release artifacts should be handled separately.

## Security Boundaries and Limitations

- The executable is a scheduled local client, not a network server; its primary
  inputs are `kpg.env`, local state, PEGELONLINE responses, and Mastodon API
  responses.
- TLS validation relies on Go's standard trust store. Certificate or DNS
  compromise, a malicious trusted CA, and compromise of the upstream services
  are outside the protections implemented by this application.
- The process has the filesystem and network privileges of its operating-system
  user. It should run as a dedicated, unprivileged account in a directory that
  other users cannot modify.
- Availability still depends on live third-party services and the scheduler.
  There are no retries, backoff, circuit breakers, or internal concurrency.
- `elwis.FetchPrognosis` is currently outside the executable path, but remains
  exported code and was included in this review.
- The test suite does not comprehensively cover posting rules, state handling,
  malformed responses, or security failure modes. Its live WSV test is not
  deterministic and may fail for environmental reasons.

## Recommended Follow-up

1. Validate WSV data before indexing or posting, propagate retrieval errors,
   and prevent invalid measurements from entering history.
2. Add request deadlines, HTTP status checks, bounded response reads, and safe
   error logging to every outbound network path.
3. Validate and restrict the Mastodon server URL before sending credentials.
4. Change state files to `0600`; use bounded, validated reads and symlink-safe,
   atomic writes in a protected working directory.
5. Harden configuration parsing and enforce owner-only configuration-file
   permissions. (`kpg.env` was corrected locally during this review.)
6. Replace fatal exits in helper packages with returned errors.
7. Add deterministic tests using local HTTP test servers for malformed,
   missing, oversized, slow, and unsuccessful responses; add tests for
   configuration, permissions, symlinks, state tampering, posting cadence, and
   duplicate emergency measurements.
8. Keep Go and module dependencies updated, install a pinned `govulncheck`
   version in CI, and rerun it on dependency and toolchain changes.
