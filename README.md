# ILR Reverse Proxy (go version)

This is a simple proxy to allow www.ilr.cornell.edu to use multiple sources for different paths of the site.

## Usage

Configuration is set via environment variables. For example:

```
LISTEN=localhost \
PORT=9700 \
DRUPAL_LATEST_URL=http://drupal.ilr.test \
DRUPAL_LEGACY_URL=http://d7.ilr.test \
./bin/proxy
```

The following environment variables are available:

`LISTEN` - The network address to listen on. Defaults to `0.0.0.0` if not set. Set this to `localhost` for testing or local development.

`PORT` - The port to listen on. Required.

`DRUPAL_LATEST_URL` - The base URL of the upstream primary Drupal site, e.g. `https://d8-edit.ilr.cornell.edu`. Required.

`DRUPAL_LEGACY_URL` - The base URL of the upstream legacy Drupal site, e.g. `https://d7-edit.ilr.cornell.edu`. Required.

## Building

```
go build -o ./bin/proxy main.go
```

## Development

You can quickly test during development with a command like this:

```
LISTEN=localhost PORT=9700 DRUPAL_LATEST_URL=http://drupal.ilr.test DRUPAL_LEGACY_URL=http://d7.ilr.test go run main.go
```

## Notes

Drupal sites use `x-forwarded-host` to configure the base_url (or equivalent). If running a Drupal site behind another reverse proxy, like we do with Caddy, we need to ensure that the correct header value is sent. See https://caddyserver.com/docs/caddyfile/options#trusted-proxies
