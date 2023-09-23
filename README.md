# KoelnPegel

This is just a small "pet project" I did to learn a little bit about the go programming language.

The program, when called, reads data on the current/last known figures about the "Pegel Köln", measuring water level and temperature of the river Rhine in Cologne. It then does - depending on a few rules - do a little post on Mastodon.

The level and temperature data is received from a public service of [PEGELONLINE](https://www.pegelonline.wsv.de/), leveraging their REST API.



## Runtime information

The actual executable is called "kpg" (KoelnPegelGo) and requires next to it a file with environment information, mainly to be able to [post on Mastodon](https://docs.joinmastodon.org/client/authorized/):

```
SERVER=https://social.cologne    # the Mastodon server to post to
CLIENT_ID=<snip>                 # client id on that server
CLIENT_SECRET=<snip>             # client secret needed to post 
ACCESS_TOKEN=<snip>              # access token to post
```

When called, the client reads historic data, decides whether to post on Mastodon or not and then updates two files: `.kpg_history` (storing the results of the last successful call to the REST API) and `.kpg_last` (storing the date and time of the last successful post to Mastodon).

If a post is made to Mastodon is decided based on quite a number of criteria - and of course these are evaluated only, if the call to the API was successful at all. These criteria do also depend on a number of predefined "thresholds" officially defined by [the City of Cologne](https://www.koeln.de/wetter/rheinpegel/).

- `GLW`: "GLW" (gleichwertiger Wasserstand) (139cm)
- `MARK_01`: "Hochwassermarke 1" (620cm)
- `MARK_02`: "Hochwassermarke 2" (830cm)
- `KATA_01`: "Rheinufertunnel läuft voll (1000cm)
- `KATA_02`: "Altstadt überflutet" (1130cm)

The algorithm then decides as follows:

 1. Check, whether current (water) level is above `KATA_02`, if yes: post, no matter what time it is
 2. Else: check, whether current (water) level is above `KATA_01`, if yes: post, if it is a full half-hour _and_ the last post has been done at least 30 minutes before
 3. Else: check, whether current (water) level is above `MARK_02`, if yes: post, if it is a full hour _and_ the last post has been done at least 60 minutes before
 4. Else: check, whether current (water) level is above `MARK_01`, if yes: post, if it is an even hour _and_ the last post has been done at least 120 minutes before
 5. Else: post, if current time is a full "4-hour" _and_ the last post has been done at least 240 minutes before

I.e., in normal conditions (no specific thresholds are reached), the program does post only all 4 hours (midnight, 4am, 8am, 12pm, 4pm, 8pm).

The actual text that is posted depends on the difference to the last known measurement (with 5cm and 10cm differences - these thresholds are a bit arbitray and derived from [BAFG](https://undine.bafg.de/rhein/zustand-aktuell/rhein_akt_WQ.html)).


## Development information

Code is split across a couple of files:

|File|Content|
|------|---|
|`kpg.go`|Main program|
|`core.go`|Routines to fetch data, do all checks and calculations and prepare text to post|
|`environment.go`|Routines to read configuration file `kpg.env`|
|`history.go`|Routines to read/write and use historic data (see file `.kpg_history`)|
|`mastopost.go`|Routine to post to Mastodon|
|`wsv/wsv.go`|Routines to call PEGELONLINE API|
|`elwis/elwis.go`|Draft routines to retrieve prognosis data - not tests and not in use yet|
|`*_test.go`|Some small unit tests|

During development, you might set the environment variable `KPG_TEST` to some value to  avoid posting to Mastodon but print the post to stdout instead.


## License

MIT License, Copyright (c) 2023 Jonas Rathert - see file `LICENSE.txt`
