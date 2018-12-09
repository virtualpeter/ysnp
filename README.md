# ysnp

builds a docker image which serves http on port 8080 by default, any requests to it are returned as a 301 redirect
to https url with no port (so 443).

run this with 8080 mapped to port 80 wherever you are serving https

it writes an apache extended log of requests to docker out.

## build:

```
$ make
```

## run:

Use docker-compose to start it up in foreground listneing to port 80

```
$ make run
```

## test:
```
$ curl -v localhost/testURI
```
  
