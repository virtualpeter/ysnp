FROM golang:1.24-alpine AS build

ARG VERSION=dev
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /ysnp ./cmd/ysnp

FROM scratch

ENV PORT=8080
ENV STATUS=301
ENV TARGET_HOST=""
ENV ALLOWED_HOSTS=""
ENV TARGET_PORT=""
ENV TARGET_PROTO="https"
ENV TARGET_PATH=""
ENV BLOCKQUERY="false"
ENV LOG="json,info"

COPY --from=build /ysnp /ysnp

EXPOSE 8080
ENTRYPOINT ["/ysnp"]
