# ysnp

HTTP to HTTPS redirector. Listens on port 8080 by default and returns a 301 to the same host and path over HTTPS (port 443).

Run it with 8080 mapped to port 80 in front of whatever serves HTTPS. Request logs are written as JSON to stdout.

## Layout

```
cmd/ysnp/            # command entrypoint
internal/server/     # redirect handler and request logger
```

## Build

```
go build -o bin/ysnp ./cmd/ysnp
```

Linux static binary in a Docker image:

```
make docker
```

## Run

Locally:

```
go run ./cmd/ysnp
```

Foreground on port 80 via Compose:

```
make compose-up
```

## Test

```
go test ./...
curl -v localhost/testURI
```

## Usage

```
  -listen string
        TCP host:port to listen on for http requests (default ":8080")
  -log string
        log flags, several allowed [debug,info,warn,error,fatal,color,nocolor,json] (default "json,info")
  -blockquery
        set if you want to block passing of request query parameters in redirect
  -status int
        http status 3xx code to return (default 301)
  -target_host string
        hardcode this domainname in redirect instead of passing on request
  -target_path string
        hardcode this path in redirect, default means use request path
  -target_port string
        port to use in redirect, default is to not have an explicit port
  -target_proto string
        protocol to redirect to, so far the only other supported option is http (default "https")
```

Flags fall back to matching environment variables when unset.

## Docker environment defaults

* `PORT=8080` (or `LISTEN=:8080`)
* `STATUS=301`
* `TARGET_HOST=""`
* `TARGET_PORT=""`
* `TARGET_PROTO="https"`
* `TARGET_PATH=""`
* `BLOCKQUERY="false"`
* `LOG="json,info"`

## Make targets

```
make build         build bin/ysnp
make test          run unit tests
make vet           run go vet
make run           go run ./cmd/ysnp
make docker        build and tag the image
make compose-up    build image and start compose
```

Image naming defaults: `REGISTRY=docker.io`, `DOCKER_ORG=virtualpete`.
