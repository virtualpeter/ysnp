
docker: ysnp
	docker build -t ysnp .
.PHONY: docker

ysnp: ysnp.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ysnp ysnp.go

run: docker
	docker-compose up
.PHONY: run