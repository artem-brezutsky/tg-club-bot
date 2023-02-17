.PHONY:
.SILENT:

build:
	go build -o ./.bin/bot cmd/bot/main.go
run:
	./.bin/bot