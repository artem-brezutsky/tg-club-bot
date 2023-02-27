.PHONY:
.SILENT:

build:
	go build -o ./.bin/bot cmd/bot/main.go

run: build run-dc-db
	./.bin/bot

build-image:
	docker-compose build

run-dc:
	docker-compose up -d

stop-dc:
	docker-compose up -d

run-dc-db:
	docker-compose up db -d