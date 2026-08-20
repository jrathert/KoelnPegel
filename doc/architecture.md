# Architecture

## Purpose

KoelnPegel is a command-line process that reports the Cologne Rhine gauge
level and water temperature to Mastodon. It obtains measurements from the
PEGELONLINE REST API and applies posting rules based on water-level thresholds,
the measurement timestamp, and the time of the previous post.

## Runtime Flow

```text
main
  |
  +-- readEnvironment("kpg.env")      -> process environment
  +-- loadHistory()                    -> .kpg_history
  +-- retrieveCurrentData()
  |     +-- wsv.QueryPegelOnline()
  |     |     +-- HTTP GET PEGELONLINE station JSON
  |     |     +-- decode JSON into WsvLevelData
  |     |     +-- select W and WT series
  |     |     +-- return Measurement
  |     +-- on API failure, return an empty Measurement
  +-- levelDifference(current, 240)    -> trend input from history
  +-- prepareStatusString(...)         -> Mastodon text
  +-- checkIfPostNow(current)
  |     +-- read .kpg_last
  |     +-- apply level/time interval rules
  |     +-- if true: postToMastodon(status)
  |     |              +-- KPG_TEST: print only
  |     |              +-- otherwise: Mastodon API
  |     |              +-- successful post: savePostTime()
  +-- append newer measurement to history
  +-- saveHistory()                    -> .kpg_history
```

The process is synchronous and has no server, scheduler, or internal retry
loop. Scheduling is expected to be provided by the caller, such as cron or a
system service.

## Components

### Root package

- `kpg.go` owns the application orchestration.
- `core.go` defines `Measurement`, water-level thresholds, trend/status text,
  posting cadence, and `.kpg_last` access.
- `environment.go` parses simple `KEY=VALUE` lines and sets process
  environment variables.
- `history.go` loads and saves the global measurement history and selects the
  historical point used for trend calculation.
- `mastopost.go` adapts the generated text to the Mastodon client. `KPG_TEST`
  makes it a dry run.

### `wsv`

`wsv.QueryPegelOnline` fetches one hard-coded PEGELONLINE station and decodes
the response. The response model expects three time series and the root
package identifies water level and temperature by short names `W` and `WT`.

### `elwis`

`elwis.FetchPrognosis` is an experimental HTML-table scraper for prognosis
data. It is not imported by the executable and is currently outside the main
runtime path.

## Data and Configuration

Configuration is read from `kpg.env` in the current working directory. The
expected variables are `SERVER`, `CLIENT_ID`, `CLIENT_SECRET`, and
`ACCESS_TOKEN`. The file is ignored by Git and must not be committed with real
credentials.

The process writes two local files:

- `.kpg_history`: indented JSON containing up to `HISTORY_LENGTH` (96)
  measurements.
- `.kpg_last`: a timestamp, encoded with `time.Time.MarshalText`, for the last
  successful Mastodon post.

These files are local operational state, not a shared database. A missing
history or last-post file is treated as an initial run. History is used to
compute the change over approximately four hours by selecting the latest
stored measurement no later than the target time.

## Posting Rules

The highest matching threshold controls the cadence:

| Level condition | Minimum interval | Timestamp condition |
|---|---:|---|
| Above `KATA_02` (1130 cm) | none | every available measurement |
| Above `KATA_01` (1000 cm) | 30 minutes | minute divisible by 30 |
| Above `MARK_02` (830 cm) | 60 minutes | minute 00 |
| Above `MARK_01` (620 cm) | 120 minutes | minute 00, even hour |
| Otherwise | 240 minutes | minute 00, hour divisible by 4 |

The implementation uses strict `>` comparisons for these posting thresholds,
while status text uses `>=` for warning messages. A successful post records
the measurement timestamp as the last-post time. A newer measurement is added
to history after the post decision, regardless of whether a post was made.

## Testing and Known Boundaries

Run `go test ./...`. Tests cover environment parsing and WSV JSON decoding, but
the WSV fetch test performs a live network request. The core decision and
history functions have limited direct test coverage.

The current implementation assumes the API response contains both `W` and
`WT` series; malformed or incomplete payloads can cause indexing failures.
Network and file errors are generally logged, and some paths return zero-value
data or stop the process. These behaviors should be preserved or deliberately
changed with focused tests when improving reliability.