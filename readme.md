# Discord Webhook Image Bot

Uploads images from `source`'s with interval defined by `cron`, each `source` has a `chance` which determines it's chance to be chosen for next run.

## Currently supported services

- ibsearch / ibsearchxxx
    - Requires `key` and `query` as `arguments` field in `source` definition.
- randomcat
    - No special `arguments`.

## How to use

`go get github.com/zet4/catsbutnotreally` it, copy `config.example.json` into a directory with the binary, rename it to `config.json` and fill in as required.


## Change log

### v0.4

- Adds a staticly compiled basic web frontend.

### v0.3

- Adds -config argument for specifying a different config file.
- Adds statistics (image sent counter) and golang stats.
- Adds pprof.
- Adds `enable_statistics`, `enable_pprof` and `web_address` to root config.
- Hot reload of the web server.

### v0.2

- Hot reload of config.
- Opt-in `embed` mode for `display` (default's to `simple`).
- MIT License added

### v0.1

- Initial release.