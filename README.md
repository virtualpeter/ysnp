# ysnp

[![CI](https://github.com/virtualpeter/ysnp/actions/workflows/ci.yml/badge.svg)](https://github.com/virtualpeter/ysnp/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A small HTTP redirector. Sit it on port 80 and send every request to the same host and path over HTTPS.

```
http://example.com/foo?q=1  →  301  →  https://example.com/foo?q=1
```

Request logs go to stdout as JSON. No dependencies beyond the Go standard library.

```mermaid
flowchart LR
  C[Client] -->|HTTP :80| Y[ysnp]
  Y -->|301 Location| C
  C -->|HTTPS :443| O[Origin]
```

## Quick start

```bash
go run ./cmd/ysnp
```

```bash
curl -sI http://localhost:8080/testURI
# HTTP/1.1 301 Moved Permanently
# Location: https://localhost/testURI
```

On port 80 with Docker Compose:

```bash
make compose-up
```

## Install

**From source**

```bash
go build -o bin/ysnp ./cmd/ysnp
```

**Docker image**

```bash
make docker
docker run --rm -p 80:8080 docker.io/virtualpete/ysnp:latest
```

Override image naming with `REGISTRY` and `DOCKER_ORG` (defaults: `docker.io` / `virtualpete`).

## Configuration

Flags win when set. Otherwise the matching environment variable is used.

| Flag | Env | Default | Description |
| --- | --- | --- | --- |
| `-listen` | `LISTEN` or `PORT` | `:8080` | Address to listen on. `PORT=8080` becomes `:8080`. |
| `-target_proto` | `TARGET_PROTO` | `https` | Scheme for the Location URL (`https` or `http`). |
| `-target_host` | `TARGET_HOST` | request host | Host to redirect to. Empty keeps the incoming `Host`. |
| `-target_port` | `TARGET_PORT` | _(none)_ | Explicit port in the Location URL. |
| `-target_path` | `TARGET_PATH` | request path | Hardcoded path. Empty keeps the request path. |
| `-blockquery` | `BLOCKQUERY` | `false` | Drop query parameters from the redirect. |
| `-status` | `STATUS` | `301` | Redirect status: `301`, `302`, `307`, or `308`. |
| `-log` | `LOG` | `json,info` | Comma-separated log options. |

### Log flags

Combine any of: `debug`, `info`, `warn`, `error`, `fatal`, `json`, `color`, `nocolor`.

```bash
./bin/ysnp -log debug,json
./bin/ysnp -log warn,nocolor
```

### Examples

Force every request onto a canonical host:

```bash
./bin/ysnp -target_host www.example.com
```

Redirect to HTTPS on 8443 and strip query strings:

```bash
./bin/ysnp -target_port 8443 -blockquery
```

Same thing in Compose:

```yaml
services:
  ysnp:
    image: docker.io/virtualpete/ysnp:latest
    ports:
      - "80:8080"
    environment:
      TARGET_HOST: www.example.com
      BLOCKQUERY: "true"
      LOG: json,info
```

## Development

```bash
make build    # bin/ysnp
make test     # go test ./...
make vet      # go vet ./...
make run      # go run ./cmd/ysnp
make docker   # build and tag the image
```

```
cmd/ysnp/            command entrypoint
internal/server/     redirect handler and request logger
```

## License

[MIT](LICENSE)
