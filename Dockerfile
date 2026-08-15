FROM golang:1.24-alpine AS build

ARG VERSION=dev
RUN adduser -D -H -u 65532 -s /sbin/nologin nonroot
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

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /ysnp /ysnp

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/ysnp"]
