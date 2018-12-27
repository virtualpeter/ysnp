# ysnp

builds a docker image which serves http on port 8080 by default, any requests to it are returned as a 301 redirect
to https url with no port (so 443).

run this with 8080 mapped to port 80 wherever you are serving https

it writes a json formatted log of requests to docker out.

## build:

To build just for local execution just use 
```
$ go build
```

but to build as linux static binary and bake into a docker image

```
$ make
```

## run:

Use docker-compose to start it up in foreground listening to port 80

```
$ make run
```

## test:
```
$ curl -v localhost/testURI
```

## Docker Release Management

To help with docker image versioning, I used a Makefile template from [https://github.com/mvanholsteijn/docker-makefile.git]

see the variable settings at the top of the Makefile to customise docker registry and username:

* REGISTRY_HOST=docker.io
* USERNAME=$(USER)

Additional documentation can be found at: [https://binx.io/blog/2017/10/07/makefile-for-docker-images/]

The Makefile has the following targets:
```
make patch-release	increments the patch release level, build and push to registry
make minor-release	increments the minor release level, build and push to registry
make major-release	increments the major release level, build and push to registry
make release		build the current release and push the image to the registry
make build		builds a new version of your Docker image and tags it
make snapshot		build from the current (dirty) workspace and pushes the image to the registry 
make check-status	will check whether there are outstanding changes
make check-release	will check whether the current directory matches the tagged release in git.
make showver		will show the current release tag based on the directory content.
```


## Command Usage:
```
Usage:
  -listen string
        TCP host:port to listen on for http requests (default ":8080")
  -log flags
        log flags, several allowed [debug,info,warn,error,fatal,color,nocolor,json]
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

## docker image environment variable defaults:
* STATUS=301
* TARGET_HOST=""
* TARGET_PORT=""
* TARGET_PROTO="https"
* TARGET_PATH=""
* BLOCKQUERY="false"
* LOG="json,info"
