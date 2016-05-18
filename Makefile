export GOPATH := $(shell pwd)
default: build

init:
	rm -f bin/server bin/main bin/mafia-server
	@cd src/main && go get

build: init
	go build -o bin/mafia-server src/main/main.go 

run: build
	@pkill ^mafia-server$ || :
	bin/mafia-server>log.txt 2>&1 &

log: run
	tail -f -n2 log.txt
