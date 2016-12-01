# Discord Webhook Image Bot

Uploads images from `source`'s with interval defined by `cron`, each `source` has a `chance` which determines it's chance to be chosen for next run.

## Currently supported

 * ibsearch 
 * * Requires `key` and `query` as `arguments` field in `source` definition.
 * randomcat
 * * No special `arguments`.

## How to use

`go get github.com/zet4/catsbutnotreally` it, copy `config.example.json` into a directory with the binary, rename it to `config.json` and fill in as required.


## Changes

### v0.2

 * Hot reload of config.
 * Opt-in `embed` mode for `display` (default's to `simple`).
 * MIT License added

### v0.1

 * Initial release.